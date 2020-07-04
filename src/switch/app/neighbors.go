package app

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/switch/storage"
	"net"
	"time"

	"github.com/danieldin95/openlan-go/src/libol"
)

type Neighbors struct {
	master Master
}

func NewNeighbors(m Master, c config.Switch) *Neighbors {
	return &Neighbors{
		master: m,
	}
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
		if arp.OpCode == libol.ArpRequest || arp.OpCode == libol.ArpReply {
			n := models.NewNeighbor(arp.SHwAddr, arp.SIpAddr, client)
			e.AddNeighbor(n)
		}
	}
	return nil
}

func (e *Neighbors) AddNeighbor(neb *models.Neighbor) {
	if n := storage.Neighbor.Get(neb.IpAddr.String()); n != nil {
		libol.Log("Neighbors.AddNeighbor: update %s.", neb)
		n.Client = neb.Client
		n.HitTime = time.Now().Unix()
	} else {
		libol.Log("Neighbors.AddNeighbor: new %s.", neb)
		storage.Neighbor.Add(neb)
	}
}

func (e *Neighbors) DelNeighbor(ipAddr net.IP) {
	libol.Info("Neighbors.DelNeighbor %s.", ipAddr)
	if n := storage.Neighbor.Get(ipAddr.String()); n != nil {
		storage.Neighbor.Del(ipAddr.String())
	}
}

func (e *Neighbors) OnClientClose(client libol.SocketClient) {
	//TODO
	libol.Info("Neighbors.OnClientClose %s.", client)
}
