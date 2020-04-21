package service

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

func (w *_network) Load(name, path string) error {
	n := &models.Network{}
	if err := libol.UnmarshalLoad(n, path); err != nil {
		libol.Error("_network.load: %s", err)
		return err
	}
	n.Name = name
	w.Add(n)
	return nil
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
	netIp := net.ParseIP(n.IpAddr)
	if netIp == nil {
		return ipStr, netmask
	}

	netIp4 := netIp.To4()
	start := binary.BigEndian.Uint32(netIp4[:4])

	for i := 0; i < n.IpRange; i++ {
		tmp := make([]byte, 4)
		binary.BigEndian.PutUint32(tmp[:4], start)

		tmpStr := net.IP(tmp).String()
		if _, ok := w.UsedAddr.GetEx(tmpStr); !ok {
			ipStr = tmpStr
			break
		}

		start += 1
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
