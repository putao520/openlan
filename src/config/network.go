package config

import "strings"

type Network struct {
	Alias     string        `json:"-"`
	Name      string        `json:"name,omitempty"`
	Provider  string        `json:"provider,omitempty"`
	Bridge    *Bridge       `json:"bridge,omitempty"`
	Subnet    *IpSubnet     `json:"subnet,omitempty"`
	OpenVPN   *OpenVPN      `json:"openvpn,omitempty"`
	Links     []*Point      `json:"links,omitempty"`
	Hosts     []HostLease   `json:"hosts,omitempty"`
	Routes    []PrefixRoute `json:"routes,omitempty"`
	Password  []Password    `json:"password,omitempty"`
	Acl       string        `json:"acl,omitempty"`
	Interface interface{}   `json:"interface,omitempty"`
	Crypt     *Crypt        `json:"crypt,omitempty"`
}

func (n *Network) Correct() {
	switch n.Provider {
	case "esp":
		port := n.Interface
		if obj, ok := port.(*ESPInterface); ok {
			obj.Correct()
			obj.Name = n.Name
		}
	case "vxlan":
		port := n.Interface
		if obj, ok := port.(*VxLANInterface); ok {
			obj.Correct()
			obj.Name = n.Name
		}
	default:
		if n.Bridge == nil {
			n.Bridge = &Bridge{}
		}
		if n.Subnet == nil {
			n.Subnet = &IpSubnet{}
		}
		br := n.Bridge
		ifAddr := ""
		br.Network = n.Name
		br.Correct()
		ifAddr = strings.SplitN(br.Address, "/", 2)[0]
		for i := range n.Routes {
			if n.Routes[i].Metric == 0 {
				n.Routes[i].Metric = 592
			}
			if n.Routes[i].NextHop == "" {
				n.Routes[i].NextHop = ifAddr
			}
			if n.Routes[i].Mode == "" {
				n.Routes[i].Mode = "snat"
			}
		}
		if n.OpenVPN != nil {
			n.OpenVPN.Network = n.Name
			obj := DefaultOpenVPN()
			n.OpenVPN.Correct(obj)
		}
	}
}
