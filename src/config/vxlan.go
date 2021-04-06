package config

import "github.com/danieldin95/openlan-go/src/libol"

type VxLANMember struct {
	Name    string `json:"name"`
	VNI     int    `json:"vni"`
	Local   string `json:"local"`
	Remote  string `json:"remote"`
	Network string `json:"network"`
	Bridge  string `json:"bridge"`
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
		if m.Local == "" {
			libol.Warn("VxLANInterface.Correct %s need local", n.Name)
			continue
		}
	}
}
