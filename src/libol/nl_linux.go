package libol

import (
	"github.com/vishvananda/netlink"
	"net"
)

func GetLocalByGw(addr string) (net.IP, error) {
	local := net.IP{}
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}
	Info("GetLocalByGw: %s", addr)
	dest := net.ParseIP(addr)
	for _, rte := range routes {
		if rte.Dst != nil && !rte.Dst.Contains(dest) {
			continue
		}
		index := rte.LinkIndex
		source := rte.Gw
		if source == nil {
			source = rte.Src
		}
		link, _ := netlink.LinkByIndex(index)
		address, _ := netlink.AddrList(link, netlink.FAMILY_V4)
		for _, ifaddr := range address {
			if ifaddr.Contains(source) {
				local = ifaddr.IP
			}
		}
	}
	return local, nil
}
