package app

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/service"
	"sync"
)

type Online struct {
	lock   sync.RWMutex
	lines  map[string]*models.Line
	worker Worker
}

func NewOnline(w Worker, c *config.VSwitch) (o *Online) {
	o = &Online{
		lines:  make(map[string]*models.Line, 1024*4),
		worker: w,
	}
	return
}

func (o *Online) OnFrame(client *libol.TcpClient, frame *libol.FrameMessage) error {
	libol.Debug("Online.OnFrame %s.", frame)
	if frame.IsControl() {
		return nil
	}

	data := frame.Data()
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("Online.OnFrame %s", err)
		return err
	}

	data = data[eth.Len:]
	if eth.IsIP4() {
		ip, err := libol.NewIpv4FromFrame(data)
		if err != nil {
			libol.Warn("Online.OnFrame %s", err)
			return err
		}
		data = data[ip.Len:]

		line := models.NewLine(eth.Type)
		line.IpSource = ip.Source
		line.IPDest = ip.Destination
		line.IpProtocol = ip.Protocol

		switch ip.Protocol {
		case libol.IPPROTO_ICMP:
			//TODO
		case libol.IPPROTO_TCP:
			tcp, err := libol.NewTcpFromFrame(data)
			if err != nil {
				libol.Warn("Online.OnFrame %s", err)
			}
			line.PortDest = tcp.Destination
			line.PortSource = tcp.Source
		case libol.IPPROTO_UDP:
			udp, err := libol.NewUdpFromFrame(data)
			if err != nil {
				libol.Warn("Online.OnFrame %s", err)
			}
			line.PortDest = udp.Destination
			line.PortSource = udp.Source
		default:
			line.PortDest = 0
			line.PortSource = 0
		}

		o.AddLine(line)
	}

	return nil
}

func (o *Online) AddLine(line *models.Line) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if _, ok := o.lines[line.String()]; !ok {
		libol.Info("Online.AddLine %s", line)
		o.lines[line.String()] = line
		service.Online.Add(line)
	}
}

func (o *Online) OnClientClose(client *libol.TcpClient) {
	//TODO
}
