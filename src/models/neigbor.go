package models

import (
	"net"
	"time"

	"github.com/danieldin95/openlan-go/src/libol"
)

type Neighbor struct {
	Client  libol.SocketClient `json:"client"`
	HwAddr  net.HardwareAddr   `json:"HwAddr"`
	IpAddr  net.IP             `json:"IpAddr"`
	NewTime int64              `json:"newTime"`
	HitTime int64              `json:"HitTime"`
}

func (e *Neighbor) String() string {
	str := e.HwAddr.String()
	str += ":" + e.IpAddr.String()
	str += ":" + e.Client.String()
	return str
}

func NewNeighbor(hwAddr net.HardwareAddr, ipAddr net.IP, client libol.SocketClient) (e *Neighbor) {
	e = &Neighbor{
		HwAddr:  hwAddr,
		IpAddr:  ipAddr,
		Client:  client,
		NewTime: time.Now().Unix(),
		HitTime: time.Now().Unix(),
	}
	return
}

func (e *Neighbor) UpTime() int64 {
	return time.Now().Unix() - e.HitTime
}
