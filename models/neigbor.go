package models

import (
	"fmt"
	"net"
	"time"

	"github.com/danieldin95/openlan-go/libol"
)

type Neighbor struct {
	Client  *libol.TcpClient `json:"Client"`
	HwAddr  net.HardwareAddr `json:"HwAddr"`
	IpAddr  net.IP           `json:"IpAddr"`
	NewTime int64            `json:"NewTime"`
	HitTime int64            `json:"HitTime"`
}

func (e *Neighbor) String() string {
	return fmt.Sprintf("%s,%s,%s", e.HwAddr, e.IpAddr, e.Client)
}

func NewNeighbor(hwAddr net.HardwareAddr, ipAddr net.IP, client *libol.TcpClient) (e *Neighbor) {
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
