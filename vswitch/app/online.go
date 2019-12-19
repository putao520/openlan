package app

import (
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/models"
	"github.com/lightstar-dev/openlan-go/service"
	"github.com/lightstar-dev/openlan-go/vswitch/api"
	"sync"
)

type Online struct {
	lock   sync.RWMutex
	lines  map[string]*models.Line
	worker api.Worker
}

func NewOnline(w api.Worker, c *config.VSwitch) (o *Online) {
	o = &Online{
		lines:  make(map[string]*models.Line, 1024*4),
		worker: w,
	}
	return
}

func (o *Online) GetOnline(name string) *models.Line {
	o.lock.RLock()
	defer o.lock.RUnlock()

	if n, ok := o.lines[name]; ok {
		return n
	}

	return nil
}

func (o *Online) ListLine() <-chan *models.Line {
	c := make(chan *models.Line, 1024)

	go func() {
		o.lock.RLock()
		defer o.lock.RUnlock()

		for _, u := range o.lines {
			c <- u
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (o *Online) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	data := frame.Data
	libol.Debug("Online.OnFrame % x.", data)

	if libol.IsControl(data) {
		return nil
	}

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
		service.Storage.SaveLine(o.worker.GetId(), line, true)
	}
}

func (o *Online) OnClientClose(client *libol.TcpClient) {
	//TODO
}


