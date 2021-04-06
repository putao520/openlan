package config

import (
	"strings"
)

type Network struct {
	Alias     string        `json:"-"`
	Name      string        `json:"name,omitempty"`
	Provider  string        `json:"provider,omitempty"`
	Bridge    Bridge        `json:"bridge,omitempty"`
	Subnet    IpSubnet      `json:"subnet,omitempty"`
	OpenVPN   *OpenVPN      `json:"openvpn,omitempty"`
	Links     []*Point      `json:"links,omitempty"`
	Hosts     []HostLease   `json:"hosts,omitempty"`
	Routes    []PrefixRoute `json:"routes,omitempty"`
	Password  []Password    `json:"password,omitempty"`
	Acl       string        `json:"acl"`
	Interface interface{}   `json:"interface,omitempty"`
	Crypt     *Crypt        `json:"crypt"`
}

func (n *Network) Correct() {
	br := &n.Bridge
	br.Network = n.Name
	br.Correct()
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
		n.OpenVPN.Correct(obj)
	}
}
