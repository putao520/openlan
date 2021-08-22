package config

import "fmt"

type FabricInterface struct {
	Mss      int              `json:"mss"`
	Name     string           `json:"name"`
	Tunnels  []*FabricTunnel  `json:"tunnels"`
	Networks []*FabricNetwork `json:"networks"`
}

func (c *FabricInterface) Correct() {
	for _, network := range c.Networks {
		network.Correct()
	}
}

type FabricTunnel struct {
	DstPort int    `json:"dport"`
	Remote  string `json:"remote"`
	Local   string `json:"local"`
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
