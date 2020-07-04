package models

import (
	"fmt"
	"net"
	"strings"
)

type Route struct {
	Prefix  string `json:"prefix"`
	NextHop string `json:"nexthop"`
}

func NewRoute(prefix string, nexthop string) (this *Route) {
	this = &Route{
		Prefix:  prefix,
		NextHop: nexthop,
	}
	return
}

func (u *Route) String() string {
	return fmt.Sprintf("%s, %s", u.Prefix, u.NextHop)
}

type Network struct {
	Name    string   `json:"name"`
	Tenant  string   `json:"tenant,omitempty"`
	IfAddr  string   `json:"ifAddr"`
	IpStart string   `json:"ipStart"`
	IpEnd   string   `json:"ipEnd"`
	Netmask string   `json:"netmask"`
	Routes  []*Route `json:"routes"`
}

func NewNetwork(name string, ifAddr string) (this *Network) {
	address := ifAddr
	netmask := "255.255.255.255"
	s := strings.SplitN(ifAddr, "/", 2)
	if len(s) == 2 {
		address = s[0]
		_, n, err := net.ParseCIDR(ifAddr)
		if err == nil {
			netmask = net.IP(n.Mask).String()
		} else {
			netmask = s[1]
		}
	}
	this = &Network{
		Name:    name,
		IfAddr:  address,
		Netmask: netmask,
	}
	return
}

func (u *Network) String() string {
	return fmt.Sprintf("%s, %s, %s, %s, %s, %s",
		u.Name, u.IfAddr, u.IpStart, u.IpEnd, u.Netmask, u.Routes)
}

func (u *Network) ParseIP(s string) {
}
