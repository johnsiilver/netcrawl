// Package statemachine provides a statemachine for processing CDP information into our network.Node. 
package statemachine

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/johnsiilver/halfpike"
	"github.com/johnsiilver/netcrawl/network"
)

// CDP is a statemachine for using a halfpike.Parser to extract data from text into a Node.
// This is used in a halfpike.Parse() and not intended to run on its own.
type CDP struct {
	node    *network.Node
	current *network.Node
	foundDevices bool
}

var deviceStart = []string{"Device", "ID:"}

// Start starts the statemachine through the text.
func (c *CDP) Start(ctx context.Context, p *halfpike.Parser) halfpike.ParseFn {
	c.node = p.Validator.(*network.Node)
	return c.findDeviceID
}

func (c *CDP) findDeviceID(ctx context.Context, p *halfpike.Parser) halfpike.ParseFn {
	c.current = &network.Node{}

	_, err := p.FindStart(deviceStart)
	if err != nil {
		if c.foundDevices {
			return nil
		}else{
			return p.Errorf("did not find any devices listed")
		}
	}
	c.foundDevices = true
	return c.findIP
}

var ipAddr = []string{"IP", "address:"}

func (c *CDP) findIP(ctx context.Context, p *halfpike.Parser) halfpike.ParseFn {
	line, until, err := p.FindUntil(ipAddr, deviceStart)
	if err != nil {
		log.Println("saw a device, but no IP listed")
		return nil
	}
	if until {
		log.Println("saw a device, but no IP listed")
		return c.findDeviceID
	}
	ipStr := halfpike.ItemJoin(line, 2, len(line.Items))

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return p.Errorf("found an IP Address: line, but couldn't decode IP(%s)", ipStr)
	}
	c.current.IP = ip
	return c.findPlatform
}

var platform = []string{"Platform:"}

func (c *CDP) findPlatform(ctx context.Context, p *halfpike.Parser) halfpike.ParseFn {
	line, until, err := p.FindUntil(platform, deviceStart)
	if err != nil {
		return p.Errorf("could not find the platform for a device")
	}
	if until {
		log.Println("saw a device, but Platform was not listed")
		return c.findDeviceID
	}

	platform := ""
	for i, item := range line.Items {
		if item.Val == "Capabilities:" {
			platform = strings.TrimRight(halfpike.ItemJoin(line, 1, i), ",")
			break
		}
	}
	if platform == "" {
		return p.Errorf("could not find the Platform in line: %s", line.Raw)
	}
	c.current.Type = platform

	return c.findInterface
}

var inter = []string{"Interface:", halfpike.Skip, "Port", "ID"}

func (c *CDP) findInterface(ctx context.Context, p *halfpike.Parser) halfpike.ParseFn {
	line, until, err := p.FindUntil(inter, deviceStart)
	if err != nil {
		return p.Errorf("could not find the interface for a device")
	}
	if until {
		log.Println("saw a device, but not what interface it was on")
		return c.findDeviceID
	}

	i := strings.TrimRight(line.Items[1].Val, ",")
	c.node.SetNeighbor(network.NodeInterface(i), c.current)
	return c.findDeviceID
}
