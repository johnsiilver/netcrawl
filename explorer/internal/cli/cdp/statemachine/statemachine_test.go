package statemachine

import (
	"context"
	"net"
	"testing"

	"github.com/johnsiilver/netcrawl/network"

	"github.com/johnsiilver/halfpike"

	"github.com/kylelemons/godebug/pretty"
)

// NOTE: Not testing failures here, but we really should.
func TestEndToEnd(t *testing.T) {
	routerOutput := `
Device ID: Switch2
Entry address(es):
  IP address: 192.168.1.243
Platform: cisco WS-C2950-12,  Capabilities: Trans-Bridge Switch
Interface: FastEthernet0/12,  Port ID (outgoing port): FastEthernet0/1
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
Device ID: Router2
Entry address(es):
  IP address: 192.168.1.240
Platform: Cisco 2621XM,  Capabilities: Switch IGMP
Interface: FastEthernet0/3,  Port ID (outgoing port): FastEthernet0/0
Holdtime : 142 sec
Version :
Cisco IOS Software, C2600 Software (C2600-ADVIPSERVICESK9-M), Version 12.3(4)T4,  RELEASE SOFTWARE (fc2)
Technical Support: http://www.cisco.com/techsupport
Copyright (c) 1986-2004 by Cisco Systems, Inc.
Compiled Thu 11-Mar-04 19:57 by eaarmas
advertisement version: 2
VTP Management Domain: ''
Duplex: full
Management address(es):
-------------------------
Device ID: RootBridge.edtetz.net
Entry address(es):
  IP address: 192.168.1.103
Platform: AIR-AP350,  Capabilities:
Interface: FastEthernet0/1,  Port ID (outgoing port): fec0
Holdtime : 131 sec
Version :
Cisco 350 Series AP 12.03T
advertisement version: 2
Duplex: full
Power drawn: 6.000 Watts
Management address(es):
	`

	ctx := context.Background()
	node := &network.Node{IP: net.ParseIP("192.168.0.1"), Type: "root node"}
	parser, err := halfpike.NewParser(routerOutput, node)
	if err != nil {
		t.Fatalf("TestEndToEnd: got err == %s", err)
	}

	sm := &CDP{}

	if err := halfpike.Parse(ctx, parser, sm.Start); err != nil {
		t.Fatalf("TestEndToEnd: got err == %s", err)
	}

	want := map[network.NodeInterface]*network.Node{
		"FastEthernet0/12": &network.Node{
			IP:   net.ParseIP("192.168.1.243"),
			Type: "cisco WS-C2950-12",
		},
		"FastEthernet0/3": &network.Node{
			IP:   net.ParseIP("192.168.1.240"),
			Type: "Cisco 2621XM",
		},
		"FastEthernet0/1": &network.Node{
			IP:   net.ParseIP("192.168.1.103"),
			Type: "AIR-AP350",
		},
	}

	if diff := pretty.Compare(want, node.Neighbors); diff != "" {
		t.Fatalf("TestEndToEnd: -want/+got:\n%s", diff)
	}
}
