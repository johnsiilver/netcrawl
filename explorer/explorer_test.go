package explorer

import (
	"context"
	"fmt"
	"net"
	"sort"
	"testing"

	"github.com/johnsiilver/netcrawl/explorer/config"
	sshCDP "github.com/johnsiilver/netcrawl/explorer/internal/cli/cdp"
	"github.com/johnsiilver/netcrawl/network"
)

var outputMap = map[string]interface{}{
	// NodeA has connections to nodeB/C/D
	"192.168.0.1": `
Device ID: nodeB
Entry address(es):
  IP address: 192.168.0.2
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/1,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
Device ID: nodeC
Entry address(es):
  IP address: 192.168.0.3
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/2,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
Device ID: nodeD
Entry address(es):
  IP address: 192.168.0.4
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/3,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
	`,
	// NodeB has connections to nodeA/C/D
	"192.168.0.2": `
Device ID: nodeA
Entry address(es):
  IP address: 192.168.0.1
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/1,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
Device ID: nodeC
Entry address(es):
  IP address: 192.168.0.3
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/2,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
Device ID: nodeD
Entry address(es):
  IP address: 192.168.0.4
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/3,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
	`,

	// NodeC has connections to nodeA/B
	"192.168.0.3": `
Device ID: nodeA
Entry address(es):
  IP address: 192.168.0.1
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/1,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
Device ID: nodeB
Entry address(es):
  IP address: 192.168.0.2
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/2,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
	`,

	// NodeD has connections to nodeA/B and a non-reachable node E
	"192.168.0.4": `
Device ID: nodeA
Entry address(es):
  IP address: 192.168.0.1
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/1,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
Device ID: nodeB
Entry address(es):
  IP address: 192.168.0.2
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/2,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
Device ID: nodeE
Entry address(es):
  IP address: 192.168.0.5
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/3,  Port ID (outgoing port): FastEthernet0/1
Holdtime : 137 sec
Version :
Cisco Internetwork Operating System Software
IOS (tm) C2950 Software (C2950-C3H2S-M), Version 12.0(5.3)WC(1), MAINTENANCE INTERIM SOFTWARE
Copyright (c) 1986-2001 by cisco Systems, Inc.
Compiled Mon 30-Apr-01 07:56 by devgoyal
advertisement version: 2
Protocol Hello:  OUI=0x00000C, Protocol ID=0x0112; payload len=27, value=00000000FFFFFFFF010121FF0000000000000006D6AC46C0FF0001
VTP Management Domain: ''
Management address(es):
-------------------------
	`,
}

func init() {
	sshCDP.FakeDialer(outputMap)
}

// Note: Tests could be much better, we don't test
func TestExplorer(t *testing.T) {
	const switchType = "cisco WS-C2950-12"
	nodeA := &network.Node{
		IP:   net.ParseIP("192.168.0.1"),
		Type: typeRoot,
	}
	nodeB := &network.Node{
		IP:   net.ParseIP("192.168.0.2"),
		Type: switchType,
	}
	nodeC := &network.Node{
		IP:   net.ParseIP("192.168.0.3"),
		Type: switchType,
	}
	nodeD := &network.Node{
		IP:   net.ParseIP("192.168.0.4"),
		Type: switchType,
	}
	nodeE := &network.Node{
		IP:   net.ParseIP("192.168.0.5"),
		Type: switchType,
	}

	nodeA.SetNeighbor("FastEthernet0/1", nodeB)
	nodeA.SetNeighbor("FastEthernet0/2", nodeC)
	nodeA.SetNeighbor("FastEthernet0/3", nodeD)

	nodeB.SetNeighbor("FastEthernet0/1", nodeA)
	nodeB.SetNeighbor("FastEthernet0/2", nodeC)
	nodeB.SetNeighbor("FastEthernet0/3", nodeD)

	nodeC.SetNeighbor("FastEthernet0/1", nodeA)
	nodeC.SetNeighbor("FastEthernet0/2", nodeB)

	nodeD.SetNeighbor("FastEthernet0/1", nodeA)
	nodeD.SetNeighbor("FastEthernet0/2", nodeB)
	nodeD.SetNeighbor("FastEthernet0/3", nodeE)

	conf := config.Config{
		SSHConn: []config.SSH{
			{User: "user", Pass: "pass"},
		},
	}
	network, err := New("192.168.0.1", conf)
	if err != nil {
		t.Fatalf("TestExplorer: New() had error: %s", err)
	}

	ctx := context.Background()
	got, err := network.Explore(ctx)
	if err != nil {
		t.Fatalf("TestExplorer: Explore() had error: %s", err)
	}

	eq := equal{seen: map[string]bool{}}
	// TODO: Really should be checking two Result objects, not two root nodes.
	if err := eq.check(nodeA, got.NetworkMap); err != nil {
		t.Fatalf("TestExplorer: %s", err)
	}

	// TODO: Check errors.
}

type equal struct {
	seen map[string]bool
}

func (e *equal) check(want, got *network.Node) error {
	if e.seen[want.IP.String()] && e.seen[got.IP.String()] {
		return nil
	}
	e.seen[want.IP.String()] = true
	e.seen[got.IP.String()] = true

	if want.IP.String() != got.IP.String() {
		return fmt.Errorf("want node is IP %s, got node is IP %s", want.IP.String(), got.IP.String())
	}

	if want.Type != got.Type {
		return fmt.Errorf("want node(%s) had type %s, got node(%s) has type %s", want.IP.String(), want.Type, got.IP.String(), got.Type)
	}

	if len(want.Neighbors) != len(got.Neighbors) {
		return fmt.Errorf("want node(%s) had neighbors %q, got node(%s) had neighbors %q",
			want.IP.String(), neighborsList(want.Neighbors), got.IP.String(), neighborsList(got.Neighbors))
	}

	if want.Error != nil || got.Error != nil {
		if want.Error != got.Error {
			return fmt.Errorf("want node(%s) had error %s, got node(%s) has error %s", want.IP.String(), want.Error, got.IP.String(), got.Error)
		}
	}

	for k, v := range want.Neighbors {
		if got.Neighbors[k] == nil {
			return fmt.Errorf("want node(%s) has neighbor %s, got node(%s) does not", want.IP.String(), k, got.IP.String())
		}

		if err := e.check(v, got.Neighbors[k]); err != nil {
			return err
		}
	}
	return nil
}

func neighborsList(m map[network.NodeInterface]*network.Node) []string {
	list := []string{}
	for k, v := range m {
		list = append(list, fmt.Sprintf("%s:%s", k, v.IP.String()))
	}

	sort.Strings(list)
	return list
}
