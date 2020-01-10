package app

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/service"
	"net"
	"sync"
	"time"

	"github.com/danieldin95/openlan-go/libol"
)

type Neighbors struct {
	lock      sync.RWMutex
	neighbors map[string]*models.Neighbor
	worker    Worker
}

func NewNeighbors(w Worker, c *config.VSwitch) (e *Neighbors) {
	e = &Neighbors{
		neighbors: make(map[string]*models.Neighbor, 1024*10),
		worker:    w,
	}
	return
}

func (e *Neighbors) OnFrame(client *libol.TcpClient, frame *libol.FrameMessage) error {
	libol.Debug("Neighbors.OnFrame %s.", frame)
	if frame.IsControl() {
		return nil
	}

	data := frame.Data()
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("Neighbors.OnFrame %s", err)
		return err
	}
	libol.Debug("Neighbors.OnFrame 0x%04x", eth.Type)
	if !eth.IsArp() {
		if eth.IsVlan() {
			//TODO
		}
		return nil
	}

	arp, err := libol.NewArpFromFrame(data[eth.Len:])
	if err != nil {
		libol.Error("Neighbors.OnFrame %s.", err)
		return nil
	}
	if arp.IsIP4() {
		if arp.OpCode == libol.ARP_REQUEST ||
			arp.OpCode == libol.ARP_REPLY {
			n := models.NewNeighbor(net.HardwareAddr(arp.SHwAddr), net.IP(arp.SIpAddr), client)
			e.AddNeighbor(n)
		}
	}

	return nil
}

func (e *Neighbors) AddNeighbor(neb *models.Neighbor) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if n, ok := e.neighbors[neb.HwAddr.String()]; ok {
		libol.Debug("Neighbors.AddNeighbor: update %s.", neb)
		n.IpAddr = neb.IpAddr
		n.Client = neb.Client
		n.HitTime = time.Now().Unix()
	} else {
		libol.Info("Neighbors.AddNeighbor: new %s.", neb)
		e.neighbors[neb.HwAddr.String()] = neb
	}

	service.Neighbor.Add(neb)
}

func (e *Neighbors) DelNeighbor(hwAddr net.HardwareAddr) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	libol.Info("Neighbors.DelNeighbor %s.", hwAddr)
	if n := e.neighbors[hwAddr.String()]; n != nil {
		service.Neighbor.Del(n.IpAddr.String())
		delete(e.neighbors, hwAddr.String())
	}
}

func (e *Neighbors) OnClientClose(client *libol.TcpClient) {
	//TODO
	libol.Info("Neighbors.OnClientClose %s.", client)
}
