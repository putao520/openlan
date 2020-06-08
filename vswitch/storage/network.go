package storage

import (
	"encoding/binary"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"net"
)

type _network struct {
	Networks   *libol.SafeStrMap
	UsedAddr   *libol.SafeStrMap
	ClientUsed *libol.SafeStrMap
}

var Network = _network{
	Networks:   libol.NewSafeStrMap(1024),
	UsedAddr:   libol.NewSafeStrMap(1024),
	ClientUsed: libol.NewSafeStrMap(1024),
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

func (w *_network) GetFreeAddr(client *libol.TcpClient, n *models.Network) (ip string, mask string) {
	if n == nil || client == nil {
		return "", ""
	}

	ipStr := ""
	netmask := n.Netmask

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
		if _, ok := w.UsedAddr.GetEx(tmpStr); !ok {
			ipStr = tmpStr
			break
		}
	}

	if ipStr != "" {
		_ = w.UsedAddr.Set(ipStr, client.Addr)
		_ = w.ClientUsed.Set(client.Addr, ipStr)
	}
	return ipStr, netmask
}

func (w *_network) FreeAddr(client *libol.TcpClient) {
	if v := w.ClientUsed.Get(client.Addr); v != nil {
		ipStr := v.(string)
		w.ClientUsed.Del(client.Addr)
		w.UsedAddr.Del(ipStr)
	}
}
