package app

import (
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/switch/storage"
	"net"
	"sync"
	"time"

	"github.com/danieldin95/openlan-go/libol"
)

type Neighbors struct {
	lock      sync.RWMutex
	neighbors map[string]*models.Neighbor
	master    Master
}

func NewNeighbors(m Master, c config.Switch) (e *Neighbors) {
	e = &Neighbors{
		neighbors: make(map[string]*models.Neighbor, 1024*10),
		master:    m,
	}
	return
}

func (e *Neighbors) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	if frame.IsControl() {
		return nil
	}
	if libol.HasLog(libol.LOG) {
		libol.Log("Neighbors.OnFrame %s.", frame)
	}
	proto, err := frame.Proto()
	if err != nil {
		libol.Warn("Neighbors.OnFrame %s", err)
		return err
	}
	eth := proto.Eth
	if !eth.IsArp() {
		return nil
	}
	arp := proto.Arp
	if arp.IsIP4() {
		if arp.OpCode == libol.ArpRequest ||
			arp.OpCode == libol.ArpReply {
			n := models.NewNeighbor(arp.SHwAddr, arp.SIpAddr, client)
			e.AddNeighbor(n)
		}
	}
	return nil
}

func (e *Neighbors) AddNeighbor(neb *models.Neighbor) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if n, ok := e.neighbors[neb.IpAddr.String()]; ok {
		libol.Log("Neighbors.AddNeighbor: update %s.", neb)
		n.IpAddr = neb.IpAddr
		n.Client = neb.Client
		n.HitTime = time.Now().Unix()
		storage.Neighbor.Update(neb)
	} else {
		libol.Log("Neighbors.AddNeighbor: new %s.", neb)
		e.neighbors[neb.IpAddr.String()] = neb
		storage.Neighbor.Add(neb)
	}
}

func (e *Neighbors) DelNeighbor(ipAddr net.IP) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	libol.Info("Neighbors.DelNeighbor %s.", ipAddr)
	if n := e.neighbors[ipAddr.String()]; n != nil {
		storage.Neighbor.Del(ipAddr.String())
		delete(e.neighbors, ipAddr.String())
	}
}

func (e *Neighbors) OnClientClose(client libol.SocketClient) {
	//TODO
	libol.Info("Neighbors.OnClientClose %s.", client)
}
