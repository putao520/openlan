package config

import "strings"

type Network struct {
	Alias    string        `json:"-"`
	Name     string        `json:"name,omitempty"`
	Bridge   Bridge        `json:"bridge,omitempty"`
	Subnet   IpSubnet      `json:"subnet,omitempty"`
	OpenVPN  *OpenVPN      `json:"openvpn,omitempty"`
	Links    []*Point      `json:"links,omitempty"`
	Hosts    []HostLease   `json:"hosts,omitempty"`
	Routes   []PrefixRoute `json:"routes,omitempty"`
	Password []Password    `json:"password,omitempty"`
}

func (n *Network) Right() {
	if n.Bridge.Name == "" {
		n.Bridge.Name = "br-" + n.Name
	}
	if n.Bridge.Provider == "" {
		n.Bridge.Provider = "linux"
	}
	if n.Bridge.IfMtu == 0 {
		n.Bridge.IfMtu = 1518
	}
	if n.Bridge.Delay == 0 {
		n.Bridge.Delay = 2
	}
	if n.Bridge.Stp == "" {
		n.Bridge.Stp = "on"
	}
	ifAddr := strings.SplitN(n.Bridge.Address, "/", 2)[0]
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
		n.OpenVPN.Right(obj)
	}
}
