package service

import (
	"encoding/binary"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"net"
)

type _network struct {
	networks   *libol.SafeStrMap
	usedAddr   *libol.SafeStrMap
	clientUsed *libol.SafeStrMap
}

var Network = _network{
	networks:   libol.NewSafeStrMap(1024),
	usedAddr:   libol.NewSafeStrMap(1024),
	clientUsed: libol.NewSafeStrMap(1024),
}

func (w *_network) Load(path string) error {
	nets := make([]*models.Network, 32)

	if err := libol.UnmarshalLoad(&nets, path); err != nil {
		libol.Error("_network.load: %s", err)
		return err
	}

	for _, net := range nets {
		w.Add(net)
	}

	return nil
}

func (w *_network) Add(n *models.Network) {
	w.networks.Set(n.Tenant, n)
}

func (w *_network) Del(name string) {
	w.networks.Del(name)
}

func (w *_network) Get(name string) *models.Network {
	if v := w.networks.Get(name); v != nil {
		return v.(*models.Network)
	}
	return nil
}

//TODO add/del route

func (w *_network) List() <-chan *models.Network {
	c := make(chan *models.Network, 128)

	go func() {
		w.networks.Iter(func(k string, v interface{}) {
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
		if _, ok := w.usedAddr.GetEx(tmpStr); !ok {
			ipStr = tmpStr
			break
		}

		start += 1
	}

	if ipStr != "" {
		w.usedAddr.Set(ipStr, client.Addr)
		w.clientUsed.Set(client.Addr, ipStr)
	}

	return ipStr, netmask
}

func (w *_network) FreeAddr(client *libol.TcpClient) {
	if v := w.clientUsed.Get(client.Addr); v != nil {
		ipStr := v.(string)
		w.clientUsed.Del(client.Addr)
		w.usedAddr.Del(ipStr)
	}
}
