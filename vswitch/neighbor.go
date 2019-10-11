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
	return time.Now().Unix() - e.NewTime
}

type Neighber struct {
	lock        sync.RWMutex
	neighbors   map[string]*Neighbor
	worker      *Worker
	EnableRedis bool
}

func NewNeighber(worker *Worker, c *Config) (e *Neighber) {
	e = &Neighber{
		neighbors:   make(map[string]*Neighbor, 1024*10),
		worker:      worker,
		EnableRedis: c.Redis.Enable,
	}
	return
}

func (e *Neighber) GetNeighbor(name string) *Neighbor {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if n, ok := e.neighbors[name]; ok {
		return n
	}

	return nil
}

func (e *Neighber) ListNeighbor() <-chan *Neighbor {
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

func (e *Neighber) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("Neighber.OnFrame % x.", frame.Data)

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
		libol.Error("Neighber.OnFrame %s.", err)
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

func (e *Neighber) AddNeighbor(neb *Neighbor) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if n, ok := e.neighbors[neb.HwAddr.String()]; ok {
		//TODO update.
		libol.Info("Neighber.AddNeighbor: update %s.", neb)
		n.IpAddr = neb.IpAddr
		n.Client = neb.Client
		n.HitTime = time.Now().Unix()
	} else {
		libol.Info("Neighber.AddNeighbor: new %s.", neb)
		n = neb
		e.neighbors[neb.HwAddr.String()] = n
	}

	e.PubNeighbor(neb, true)
}

func (e *Neighber) DelNeighbor(hwAddr net.HardwareAddr) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	libol.Info("Neighber.DelNeighbor %s.", hwAddr)
	if n := e.neighbors[hwAddr.String()]; n != nil {
		e.PubNeighbor(n, false)
		delete(e.neighbors, hwAddr.String())
	}
}

func (e *Neighber) OnClientClose(client *libol.TcpClient) {
	//TODO
	libol.Info("Neighber.OnClientClose %s.", client)
}

func (e *Neighber) PubNeighbor(neb *Neighbor, isadd bool) {
	if !e.EnableRedis {
		return
	}

	key := fmt.Sprintf("neighbor:%s", strings.Replace(neb.HwAddr.String(), ":", "-", -1))
	value := map[string]interface{}{
		"hwAddr":  neb.HwAddr.String(),
		"ipAddr":  neb.IpAddr.String(),
		"remote":  neb.Client.String(),
		"newTime": neb.NewTime,
		"hitTime": neb.HitTime,
		"active":  isadd,
	}

	if err := e.worker.Redis.HMSet(key, value); err != nil {
		libol.Error("Neighber.PubNeighbor hset %s", err)
	}
}
