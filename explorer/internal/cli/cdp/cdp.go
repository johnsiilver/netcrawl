// Package CDP provides a method for doing neighbor discovery via a command line SSH CDP command.
package cdp

import (
	"context"
	"fmt"
	"net"

	"github.com/johnsiilver/halfpike"
	"github.com/johnsiilver/netcrawl/explorer/internal/cli/cdp/statemachine"
	"github.com/johnsiilver/netcrawl/network"
	"golang.org/x/crypto/ssh"
)

// Discover will try to discover a node via CDP via an SSH CLI session.
type Discover struct {
	configs []*ssh.ClientConfig
}

// New is the constructor for Discover.
func New(configs []*ssh.ClientConfig) (*Discover, error) {
	return &Discover{configs: configs}, nil
}

// Node logs into node.IP and runs CDP neighbor discovery and fills out our Neighbors.
func (d *Discover) Node(ctx context.Context, node *network.Node) error {
	var b []byte
	var err error

	for _, conf := range d.configs {
		b, err = d.runCDPNeighbor(node.IP, conf)
		if err == nil {
			break
			return fmt.Errorf("could not login to node %s: %s", node.IP.String(), err)
		}
	}
	if err != nil {
		return fmt.Errorf("could not login to node(%s) with any provided user/password, last error was: %s", node.IP.String(), err)
	}

	parser, err := halfpike.NewParser(string(b), node)
	if err != nil {
		return fmt.Errorf("problems making parser for node %s output: %s", node.IP.String(), err)
	}

	sm := &statemachine.CDP{}

	if err := halfpike.Parse(ctx, parser, sm.Start); err != nil {
		return err
	}
	return nil
}

func (d *Discover) runCDPNeighbor(nodeIP net.IP, config *ssh.ClientConfig) ([]byte, error) {
	cli, err := dialer(nodeIP.String(), config)
	if err != nil {
		return nil, fmt.Errorf("could not dial node: %s", err)
	}
	defer cli.conn().close()

	session, err := cli.newSession()
	if err != nil {
		return nil, fmt.Errorf("could not create session: %s", err)
	}
	defer session.close()

	b, err := session.combinedOutput("show cdp neighbors detail")
	if err != nil {
		return nil, fmt.Errorf("problem executing 'show cdp neighbors detail': %s", err)
	}
	return b, nil
}
