package config

import (
	"context"
	"fmt"
	"time"

	sshCDP "github.com/johnsiilver/netcrawl/explorer/internal/cli/cdp"
	"github.com/johnsiilver/netcrawl/network"
	"golang.org/x/crypto/ssh"
)

type Discover interface {
	Node(ctx context.Context, node *network.Node) error
}

// Config represents a configuration file for netcrawl.
type Config struct {
	// SSHConn provides a list of possible ssh configurations that would
	// allow connection to the device.
	SSHConn []SSH
	// SNMPConn []SNMP
}

func (c Config) Discoveries() ([]Discover, error) {
	var discNodes []Discover

	discs, err := c.sshDiscovery()
	if err != nil {
		return nil, err
	}

	discNodes = append(discNodes, discs...)

	return discNodes, nil
}

func (c Config) sshDiscovery() ([]Discover, error) {
	var discNodes []Discover
	var sshConfigs []*ssh.ClientConfig

	for _, sshConf := range c.SSHConn {
		config := &ssh.ClientConfig{
			User: sshConf.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(sshConf.Pass),
			},
			Timeout: 5 * time.Second,
		}
		sshConfigs = append(sshConfigs, config)
	}

	if len(sshConfigs) > 0 {
		disc, err := sshCDP.New(sshConfigs)
		if err != nil {
			return nil, fmt.Errorf("problems setting up SSH CDP discovery: %s", err)
		}
		discNodes = append(discNodes, disc)
	}
	return discNodes, nil
}

// SSH provides an SSH configuration for connecting to a device.
type SSH struct {
	User string
	Pass string
}

/*

type SNMP interface{
	isSNMP()
}

type SNMPv2 struct {
	Community string
	Port int
}

func (SNMPv2) isSNMP(){}

type SNMPv3 struct {

}
func (SNMPv3) isSNMP(){}
*/
