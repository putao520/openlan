package vswitch

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lightstar-dev/openlan-go/libol"
)

type Neighbors struct {
	lock        sync.RWMutex
	neighbors   map[string]*Neighbor
	worker      WorkerApi
	EnableRedis bool
}

func NewNeighbors(api WorkerApi, c *Config) (e *Neighbors) {
	e = &Neighbors{
		neighbors:   make(map[string]*Neighbor, 1024*10),
		worker:      api,
		EnableRedis: c.Redis.Enable,
	}
	return
}

func (e *Neighbors) GetNeighbor(name string) *Neighbor {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if n, ok := e.neighbors[name]; ok {
		return n
	}

	return nil
}

func (e *Neighbors) ListNeighbor() <-chan *Neighbor {
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

func (e *Neighbors) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("Neighbors.OnFrame % x.", frame.Data)

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
		libol.Error("Neighbors.OnFrame %s.", err)
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

func (e *Neighbors) AddNeighbor(neb *Neighbor) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if n, ok := e.neighbors[neb.HwAddr.String()]; ok {
		//TODO update.
		libol.Info("Neighbors.AddNeighbor: update %s.", neb)
		n.IpAddr = neb.IpAddr
		n.Client = neb.Client
		n.HitTime = time.Now().Unix()
	} else {
		libol.Info("Neighbors.AddNeighbor: new %s.", neb)
		n = neb
		e.neighbors[neb.HwAddr.String()] = n
	}

	e.PubNeighbor(neb, true)
}

func (e *Neighbors) DelNeighbor(hwAddr net.HardwareAddr) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	libol.Info("Neighbors.DelNeighbor %s.", hwAddr)
	if n := e.neighbors[hwAddr.String()]; n != nil {
		e.PubNeighbor(n, false)
		delete(e.neighbors, hwAddr.String())
	}
}

func (e *Neighbors) OnClientClose(client *libol.TcpClient) {
	//TODO
	libol.Info("Neighbors.OnClientClose %s.", client)
}

func (e *Neighbors) PubNeighbor(neb *Neighbor, isadd bool) {
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

	if err := e.worker.GetRedis().HMSet(key, value); err != nil {
		libol.Error("Neighbors.PubNeighbor hset %s", err)
	}
}
