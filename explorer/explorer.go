// Package explorer provides facilities to explorer a network by connecting to a root node via
// a user/password and issuing CDP commands and then logging into its neighbors in parallel.
// This mostly just tries to Go version the Python script, with enhancements (like this will
// be way faster in a large network due to concurrency).
//
// Note1: There really should be multiple explorers here. I'd probably start with SNMP. If that
// failed I might try device login with a user/pass/ssh key list. Enterprise network engineers
// have a bad tendancy to leave SNMP set to "public" or something easily guessable so they
// can run solarwinds or nagios easily.  Yes, SNMP is terrible,
// but it can provide structured data for these old routers you can't export JSON from or do streaming
// telemetry.
//
// Note2: CDP or LLDP could be used.
//
// Note3: Routing protocols can give you enough information to detect neighbors. While not giving
// you the discovery that CDP/LLDP can, you can then just use that info to query the neighborts
// (OSPF/ISIS/...)
//
// Note4: Not that I've dealt with pure switches in a while, but if your ARP table shows something, then
// you have a device there. If there is a link light, something has a mac address at least in a table. Using
// the OID you can get an idea of what is hanging off there even if you don't know what it is.
//
// Note5: I'm sure this doesn't deal with a bunch of various interface options. And I no longer do
// network automation, so I don't have a bunch of devices to try this on.  Basically, this is a REALLY
// junior effort.  Just as a mental exercise while I didn't want to work on other things I should.
package explorer

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/johnsiilver/halfpike"
	"golang.org/x/crypto/ssh"
	"github.com/johnsiilver/netcrawl/internal/statemachine"
	"github.com/johnsiilver/netcrawl/network"
)

// LoginDeny is used to log nodes we could not connect to.
type LoginDeny struct {
	// IP is the IP of the node.
	IP net.IP
	// Err is the error we got.
	Err error
}

type Results struct {
	NetworkMap *network.Node
	LoginDeny []LoginDeny
	ParseErrors []error
}

// Network is used to explorer the network
type Network struct {
	root       *network.Node
	user, pass string

	error      error
	loginDeny  []LoginDeny
	parseError []error
	seen       map[string]*network.Node // keys are net.IP.String()

	mu sync.Mutex
	wg sync.WaitGroup
}

const typeRoot = "RootNode"

// New is the constructor for Network.
func New(root, user, pass string) (*Network, error) {
	ip := net.ParseIP(root)
	if ip == nil {
		ips, err := net.LookupIP(root)
		if err != nil {
			return nil, fmt.Errorf("root node %s was not an IP and could not be found in DNS", root)
		}
		ip = ips[0]
	}
	n := &network.Node{IP: ip, Type: typeRoot}
	return &Network{
		root: n,
		seen: map[string]*network.Node{ip.String(): n},
		user: user,
		pass: pass,
	}, nil
}

func (e *Network) Explore(ctx context.Context) (Results, error) {
	e.wg.Add(1)
	e.processNode(ctx, e.root, nil, "")

	e.wg.Wait()

	if e.error != nil {
		return Results{}, e.error
	}

	return Results{
		NetworkMap: e.root,
		LoginDeny: e.loginDeny,
		ParseErrors: e.parseError,
	}, nil
}

func (e *Network) seenNode(n *network.Node) *network.Node {
	e.mu.Lock()
	defer e.mu.Unlock()

	seen := e.seen[n.IP.String()]
	if seen == nil {
		e.seen[n.IP.String()] = n
		return nil
	}
	return seen
}

func (e *Network) processNode(ctx context.Context, node, parent *network.Node, inter network.NodeInterface) {
	defer e.wg.Done()

	b, err := e.runCDPNeighbor(node.IP)
	if err != nil {
		if parent == nil {
			e.error = fmt.Errorf("could not connect to root node: %s", err)
			return
		}
		e.loginDeny = append(e.loginDeny, LoginDeny{node.IP, err})
		return
	}

	parser, err := halfpike.NewParser(string(b), node)
	if err != nil {
		e.error = err
		return
	}

	sm := &statemachine.CDP{}

	if err := halfpike.Parse(ctx, parser, sm.Start); err != nil {
		if parent == nil {
			e.error = fmt.Errorf("problems reading CDP from root node")
			return
		}
		e.parseError = append(e.parseError, err)
		return
	}
	if parent != nil {
		parent.SetNeighbor(inter, node)
	}

	e.wg.Add(1)
	go e.walkChildren(ctx, node, inter)
}

func (e *Network) walkChildren(ctx context.Context, parent *network.Node, inter network.NodeInterface) {
	defer e.wg.Done()

	for inter, child := range parent.Neighbors {
		if seen := e.seenNode(child); seen != nil {
			// The node information here will be incomplete (missing Neighbors).
			// This completes it.
			parent.SetNeighbor(inter, seen)
			continue
		}

		e.wg.Add(1)
		go e.processNode(ctx, child, parent, inter)
	}
}

func (e *Network) runCDPNeighbor(nodeIP net.IP) ([]byte, error) {
	config := ssh.ClientConfig{
		User: e.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(e.pass),
		},
		Timeout: 5 * time.Second,
	}

	cli, err := dialer(nodeIP.String(), &config)
	if err != nil {
		return nil, fmt.Errorf("could not connect to root node at: %s", nodeIP)
	}
	defer cli.conn().close()

	session, err := cli.newSession()
	if err != nil {
		return nil, fmt.Errorf("could not create session to node(%s): %s", nodeIP, err)
	}
	defer session.close()

	b, err := session.combinedOutput("show cdp neighbors detail")
	if err != nil {
		return nil, fmt.Errorf("problem executing 'show cdp neighbors detail' on node(%s): %s", nodeIP, err)
	}
	return b, nil
}

// List provides a method for walking the network.Node tree and returning a list of all Nodes
// without falling into a recursive loop.
type List struct {
	seen map[string]bool
	list []*network.Node
}

// List turns the tree into a slice.
func (l *List) List(n *network.Node) []*network.Node{
	l.seen = map[string]bool{}
	l.list = []*network.Node{}
	l.walk(n)
	return l.list
}

func (l *List) walk(n *network.Node) {
	if l.seen[n.IP.String()]{
		return
	}
	l.seen[n.IP.String()] = true
	l.list = append(l.list, n)
}
