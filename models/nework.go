package models

import "fmt"

type Route struct {
	Prefix  string `json:"prefix"`
	Nexthop string `json:"nexthop"`
}

func NewRoute(prefix string, nexthop string) (this *Route) {
	this = &Route{
		Prefix:  prefix,
		Nexthop: nexthop,
	}
	return
}

func (u *Route) String() string {
	return fmt.Sprintf("%s, %s", u.Prefix, u.Nexthop)
}

type Network struct {
	Name    string   `json:"name"`
	IfAddr  string   `json:"ifAddr"`
	IpAddr  string   `json:"ipAddr"`
	IpRange int      `json:"ipRange"`
	Netmask string   `json:"netmask"`
	Routes  []*Route `json:"routes"`
}

func NewNetwork(name string, ifAddr string) (this *Network) {
	this = &Network{
		Name:   name,
		IfAddr: ifAddr,
	}
	return
}

func (u *Network) String() string {
	return fmt.Sprintf("%s, %s, %s, %d, %s, %s",
		u.Name, u.IfAddr, u.IpAddr, u.IpRange, u.Netmask, u.Routes)
}

func (u *Network) ParseIP(s string) {

}
