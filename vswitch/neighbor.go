package vswitch

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lightstar-dev/openlan-go/libol"
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

func NewNeighbor(hwaddr net.HardwareAddr, ipaddr net.IP, client *libol.TcpClient) (e *Neighbor) {
	e = &Neighbor{
		HwAddr:  hwaddr,
		IpAddr:  ipaddr,
		Client:  client,
		NewTime: time.Now().Unix(),
		HitTime: time.Now().Unix(),
	}

	return
}

func (e *Neighbor) UpTime() int64 {
	return time.Now().Unix() - e.NewTime
}

type Neighborer struct {
	lock        sync.RWMutex
	neighbors   map[string]*Neighbor
	wroker      *Worker
	EnableRedis bool
}

func NewNeighborer(wroker *Worker, c *Config) (e *Neighborer) {
	e = &Neighborer{
		neighbors:   make(map[string]*Neighbor, 1024*10),
		wroker:      wroker,
		EnableRedis: c.Redis.Enable,
	}
	return
}

func (e *Neighborer) GetNeighbor(name string) *Neighbor {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if n, ok := e.neighbors[name]; ok {
		return n
	}

	return nil
}

func (e *Neighborer) ListNeighbor() <-chan *Neighbor {
	c := make(chan *Neighbor, 128)

	go func() {
		e.lock.RLock()
		defer e.lock.RUnlock()

		for _, u := range e.neighbors {
			c <- u
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (e *Neighborer) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("Neighborer.OnFrame % x.", frame.Data)

	if libol.IsInst(frame.Data) {
		return nil
	}

	eth, err := libol.NewEtherFromFrame(frame.Data)
	if err != nil {
		libol.Warn("PointCmd.onArp %s", err)
		return err
	}
	if !eth.IsArp() {
		if eth.IsVlan() {
			//TODO
		}
		return nil
	}

	arp, err := libol.NewArpFromFrame(frame.Data[eth.Len:])
	if err != nil {
		libol.Error("Neighborer.OnFrame %s.", err)
		return nil
	}
	if arp.IsIP4() {
		if arp.OpCode == libol.ARP_REQUEST ||
			arp.OpCode == libol.ARP_REPLY {
			n := NewNeighbor(net.HardwareAddr(arp.SHwAddr), net.IP(arp.SIpAddr), client)
			e.AddNeighbor(n)
		}
	}

	return nil
}

func (e *Neighborer) AddNeighbor(neb *Neighbor) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if n, ok := e.neighbors[neb.HwAddr.String()]; ok {
		//TODO update.
		libol.Info("Neighborer.AddNeighbor: update %s.", neb)
		n.IpAddr = neb.IpAddr
		n.Client = neb.Client
		n.HitTime = time.Now().Unix()
	} else {
		libol.Info("Neighborer.AddNeighbor: new %s.", neb)
		n = neb
		e.neighbors[neb.HwAddr.String()] = n
	}

	e.PubNeighbor(neb, true)
}

func (e *Neighborer) DelNeighbor(hwaddr net.HardwareAddr) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	libol.Info("Neighborer.DelNeighbor %s.", hwaddr)
	if n := e.neighbors[hwaddr.String()]; n != nil {
		e.PubNeighbor(n, false)
		delete(e.neighbors, hwaddr.String())
	}
}

func (e *Neighborer) OnClientClose(client *libol.TcpClient) {
	//TODO
	libol.Info("Neighborer.OnClientClose %s.", client)
}

func (e *Neighborer) PubNeighbor(neb *Neighbor, isadd bool) {
	if !e.EnableRedis {
		return
	}

	key := fmt.Sprintf("neighbor:%s", strings.Replace(neb.HwAddr.String(), ":", "-", -1))
	value := map[string]interface{}{
		"hwaddr":  neb.HwAddr.String(),
		"ipaddr":  neb.IpAddr.String(),
		"remote":  neb.Client.String(),
		"newtime": neb.NewTime,
		"hittime": neb.HitTime,
		"actived": isadd,
	}

	if err := e.wroker.Redis.HMSet(key, value); err != nil {
		libol.Error("Neighborer.PubNeighbor hset %s", err)
	}
}
