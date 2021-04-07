package libol

import (
	"github.com/vishvananda/netlink"
	"net"
)

func GetAddrByGw() (net.IP, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	for _, route := range routes {
		if route.Dst != nil {
			continue
		}
		index := route.LinkIndex
		gateway := route.Gw
		link, _ := netlink.LinkByIndex(index)
		adders, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			return nil, err
		}
		for _, addr := range adders {
			if addr.Contains(gateway) {
				return addr.IP, nil
			}
		}
	}
	return nil, NewErr("notFound")
}
