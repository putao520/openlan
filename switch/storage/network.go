package storage

import (
	"encoding/binary"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/switch/schema"
	"net"
)

type _network struct {
	Networks *libol.SafeStrMap
	AddrUUID *libol.SafeStrStr
	UUIDAddr *libol.SafeStrStr
}

var Network = _network{
	Networks: libol.NewSafeStrMap(1024),
	AddrUUID: libol.NewSafeStrStr(1024),
	UUIDAddr: libol.NewSafeStrStr(1024),
}

func (w *_network) Add(n *models.Network) {
	libol.Debug("_network.Add %v", *n)
	_ = w.Networks.Set(n.Name, n)
}

func (w *_network) Del(name string) {
	libol.Debug("_network.Del %s", name)
	w.Networks.Del(name)
}

func (w *_network) Get(name string) *models.Network {
	if v := w.Networks.Get(name); v != nil {
		return v.(*models.Network)
	}
	return nil
}

//TODO add/del route

func (w *_network) List() <-chan *models.Network {
	c := make(chan *models.Network, 128)

	go func() {
		w.Networks.Iter(func(k string, v interface{}) {
			c <- v.(*models.Network)
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (w *_network) ListLease() <-chan *schema.Lease {
	c := make(chan *schema.Lease, 128)

	go func() {
		w.UUIDAddr.Iter(func(k string, v string) {
			c <- &schema.Lease{
				UUID:    k,
				Address: v,
				Client:  Point.GetAddr(k),
			}
		})
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (w *_network) GetFreeAddr(uuid string, n *models.Network) (ip string, mask string) {
	if n == nil || uuid == "" {
		return "", ""
	}

	ipStr := ""
	netmask := n.Netmask

	if addr, ok := w.UUIDAddr.GetEx(uuid); ok {
		return addr, netmask
	}

	sIp := net.ParseIP(n.IpStart)
	eIp := net.ParseIP(n.IpEnd)
	if sIp == nil || eIp == nil {
		return ipStr, netmask
	}

	start := binary.BigEndian.Uint32(sIp.To4()[:4])
	end := binary.BigEndian.Uint32(eIp.To4()[:4])
	for i := start; i <= end; i++ {
		tmp := make([]byte, 4)
		binary.BigEndian.PutUint32(tmp[:4], i)
		tmpStr := net.IP(tmp).String()
		if _, ok := w.AddrUUID.GetEx(tmpStr); !ok {
			ipStr = tmpStr
			break
		}
	}

	if ipStr != "" {
		_ = w.AddrUUID.Set(ipStr, uuid)
		_ = w.UUIDAddr.Set(uuid, ipStr)
	}
	return ipStr, netmask
}

func (w *_network) FreeAddr(uuid string) {
	if addr, ok := w.UUIDAddr.GetEx(uuid); ok {
		w.UUIDAddr.Del(uuid)
		w.AddrUUID.Del(addr)
	}
}
