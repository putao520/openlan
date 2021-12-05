package config

import "fmt"

type FabricSpecifies struct {
	Mss      int              `json:"tcpMss,omitempty"`
	Fragment bool             `json:"fragment"`
	Driver   string           `json:"driver,omitempty"`
	Name     string           `json:"name"`
	Tunnels  []*FabricTunnel  `json:"tunnels"`
	Networks []*FabricNetwork `json:"networks"`
}

func (c *FabricSpecifies) Correct() {
	for _, network := range c.Networks {
		network.Correct()
	}
	for _, tun := range c.Tunnels {
		tun.Correct()
		if tun.DstPort == 0 {
			if c.Driver == "stt" {
				tun.DstPort = 7471
			} else {
				tun.DstPort = 4789
			}
		}
	}
}

type FabricTunnel struct {
	DstPort uint32 `json:"dport"`
	Remote  string `json:"remote"`
	Local   string `json:"local,omitempty" yaml:"local,omitempty"`
}

func (c *FabricTunnel) Correct() {
}

type FabricNetwork struct {
	Vni     uint32         `json:"vni"`
	Bridge  string         `json:"bridge"`
	Outputs []FabricOutput `json:"outputs"`
}

func (c *FabricNetwork) Correct() {
	if c.Bridge == "" {
		c.Bridge = fmt.Sprintf("br-%08x", c.Vni)
	}
}

type FabricOutput struct {
	Vlan      int    `json:"vlan"`
	Interface string `json:"interface"`
}
