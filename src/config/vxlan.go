package config

import (
	"fmt"
)

type VxLANMember struct {
	Name    string `json:"name,omitempty"`
	VNI     int    `json:"vni"`
	Local   string `json:"local,omitempty"`
	Remote  string `json:"remote"`
	Network string `json:"network,omitempty"`
	Bridge  string `json:"bridge,omitempty"`
	Port    int    `json:"port,omitempty"`
}

func (m *VxLANMember) Correct() {
	if m.Name == "" {
		m.Name = fmt.Sprintf("vni%d", m.VNI)
	}
}

type VxLANInterface struct {
	Name    string         `json:"name"`
	Local   string         `json:"local,omitempty"`
	Bridge  string         `json:"bridge,omitempty"`
	Address string         `json:"address,omitempty"`
	Members []*VxLANMember `json:"members"`
}

func (n *VxLANInterface) Correct() {
	for _, m := range n.Members {
		if m.Local == "" {
			m.Local = n.Local
		}
		if m.Bridge == "" {
			m.Bridge = n.Bridge
		}
		m.Correct()
	}
}
