package config

import "fmt"

type FabricSpecifies struct {
	Mss      int              `json:"tcpMss"`
	Name     string           `json:"name"`
	Tunnels  []*FabricTunnel  `json:"tunnels"`
	Networks []*FabricNetwork `json:"networks"`
}

func (c *FabricSpecifies) Correct() {
	if c.Mss == 0 {
		c.Mss = 1332
	}
	for _, network := range c.Networks {
		network.Correct()
	}
	for _, tunnel := range c.Tunnels {
		tunnel.Correct()
	}
}

type FabricTunnel struct {
	DstPort uint32 `json:"dport"`
	Remote  string `json:"remote"`
	Local   string `json:"local,omitempty" yaml:"local,omitempty"`
}

func (c *FabricTunnel) Correct() {
	if c.DstPort == 0 {
		c.DstPort = 4789
	}
}

type FabricNetwork struct {
	Vni     uint32         `json:"vni"`
	Bridge  string         `json:"bridge"`
	Outputs []FabricOutput `json:"outputs"`
}

func (c *FabricNetwork) Correct() {
	if c.Bridge == "" {
		c.Bridge = fmt.Sprintf("br-%x", c.Vni)
	}
}

type FabricOutput struct {
	Vlan      int    `json:"vlan"`
	Interface string `json:"interface"`
}
