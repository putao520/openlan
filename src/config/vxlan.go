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
	Members []*VxLANMember `json:"members"`
}

func (n *VxLANInterface) Correct() {
	for _, m := range n.Members {
		if m.Local == "" {
			m.Local = n.Local
		}
		m.Correct()
	}
}
