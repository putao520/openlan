package app

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/service"
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

func NewNeighbors(m Master, c config.VSwitch) (e *Neighbors) {
	e = &Neighbors{
		neighbors: make(map[string]*models.Neighbor, 1024*10),
		master:    m,
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

	if n, ok := e.neighbors[neb.IpAddr.String()]; ok {
		libol.Debug("Neighbors.AddNeighbor: update %s.", neb)
		n.IpAddr = neb.IpAddr
		n.Client = neb.Client
		n.HitTime = time.Now().Unix()
		service.Neighbor.Update(neb)
	} else {
		libol.Info("Neighbors.AddNeighbor: new %s.", neb)
		e.neighbors[neb.IpAddr.String()] = neb
		service.Neighbor.Add(neb)
	}
}

func (e *Neighbors) DelNeighbor(ipAddr net.IP) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	libol.Info("Neighbors.DelNeighbor %s.", ipAddr)
	if n := e.neighbors[ipAddr.String()]; n != nil {
		service.Neighbor.Del(ipAddr.String())
		delete(e.neighbors, ipAddr.String())
	}
}

func (e *Neighbors) OnClientClose(client *libol.TcpClient) {
	//TODO
	libol.Info("Neighbors.OnClientClose %s.", client)
}
