package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
)

func main() {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		panic(err)
	}
	for _, route := range routes {
		// fmt.Println(route)
		if route.Dst != nil {
			continue
		}
		ifIndex := route.LinkIndex
		gateway := route.Gw
		link, _ := netlink.LinkByIndex(ifIndex)
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			if addr.Contains(gateway) {
				fmt.Println(addr.IP)
			}
		}
	}
}
