package config

import "strings"

type Network struct {
	File      string        `json:"file"`
	Alias     string        `json:"-" yaml:"-"`
	Name      string        `json:"name,omitempty" yaml:"name"`
	Provider  string        `json:"provider,omitempty" yaml:"provider"`
	Bridge    *Bridge       `json:"bridge,omitempty" yaml:"bridge,omitempty"`
	Subnet    *IpSubnet     `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	OpenVPN   *OpenVPN      `json:"openvpn,omitempty" yaml:"openvpn,omitempty"`
	Links     []*Point      `json:"links,omitempty" yaml:"links,omitempty"`
	Hosts     []HostLease   `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Routes    []PrefixRoute `json:"routes,omitempty" yaml:"routes,omitempty"`
	Password  []Password    `json:"password,omitempty" yaml:"password,omitempty"`
	Acl       string        `json:"acl,omitempty" yaml:"acl,omitempty"`
	Specifies interface{}   `json:"specifies,omitempty" yaml:"specifies,omitempty"`
	Crypt     *Crypt        `json:"crypt,omitempty" yaml:"crypt,omitempty"`
}

func (n *Network) Correct() {
	switch n.Provider {
	case "esp":
		spec := n.Specifies
		if obj, ok := spec.(*ESPSpecifies); ok {
			obj.Correct()
			obj.Name = n.Name
		}
	case "vxlan":
		spec := n.Specifies
		if obj, ok := spec.(*VxLANSpecifies); ok {
			obj.Correct()
			obj.Name = n.Name
		}
		br := n.Bridge
		if br != nil {
			br.Network = n.Name
			br.Correct()
			// 28 [udp] - 8 [esp] -
			// 28 [udp] - 8 [vxlan] -
			// 14 [ethernet] - tcp [40] - 1332 [mss] -
			// 42 [padding] ~= variable 30-45
			if br.Mss == 0 {
				br.Mss = 1332
			}
		}
	case "fabric":
		spec := n.Specifies
		if obj, ok := spec.(*FabricSpecifies); ok {
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
