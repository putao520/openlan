package schema

import (
	"net"
)

type Neighbor struct {
	Client  string           `json:"client"`
	HwAddr  net.HardwareAddr `json:"hwAddr"`
	IpAddr  net.IP           `json:"ipAddr"`
	NewTime int64            `json:"newTime"`
	HitTime int64            `json:"hitTime"`
}
