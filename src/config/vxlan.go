package config

import (
	"fmt"
)

type VxLANMember struct {
	Name    string `json:"name"`
	VNI     int    `json:"vni"`
	Local   string `json:"local"`
	Remote  string `json:"remote"`
	Network string `json:"network"`
	Bridge  string `json:"bridge"`
	Port    int    `json:"port"`
}

func (m *VxLANMember) Correct() {
	if m.Name == "" {
		m.Name = fmt.Sprintf("vxlan-%d", m.VNI)
	}
}

type VxLANInterface struct {
	Name    string         `json:"name"`
	Local   string         `json:"local"`
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
