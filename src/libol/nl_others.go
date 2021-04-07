// +build !linux

package libol

import "net"

func GetAddrByGw() (net.IP, error) {
	return nil, NewErr("GetAddrByGw notSupport")
}
