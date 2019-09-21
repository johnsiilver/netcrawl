package network

import (
	"fmt"
	"net"
	"sync"
)

// NodeInterface is a vendor specific network interface now.
type NodeInterface string

// Node represents a network ndoe.
type Node struct {
	// IP is the IP the Node is connected with.
	IP net.IP
	// Type is the type of device.  In a real version of this, this should be an enumerator
	// an this should probably based on protocol buffers.
	Type string

	// Neighbors is a set of Interaces that connect to a Neighbor.
	Neighbors map[NodeInterface]*Node

	// Error indicates errors associates with logging into the node or parsing CDP.
	Error error

	mu sync.Mutex
}

// SetNeighbor sets a Neighbor at inter to node.
func (n *Node) SetNeighbor(inter NodeInterface, node *Node) {
	n.mu.Lock()
	if n.Neighbors == nil {
		n.Neighbors = map[NodeInterface]*Node{}
	}
	n.Neighbors[inter] = node
	n.mu.Unlock()
}

// Validate vlaidates that this Node is valid.
func (n *Node) Validate() error {
	if n.IP == nil {
		return fmt.Errorf("IP was not set")
	}
	if n.Type == "" {
		return fmt.Errorf("Type was empty")
	}
	for k, v := range n.Neighbors {
		if v == nil {
			return fmt.Errorf("Node had nil Neighbor on interfade %s", k)
		}
	}
	return nil
}
