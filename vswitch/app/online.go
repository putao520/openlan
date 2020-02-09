package app

import (
	"container/list"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/service"
	"sync"
	"time"
)

type Online struct {
	max       int
	lock      sync.RWMutex
	lines     map[string]*models.Line
	lineList  *list.List
	worker    Worker
}

func NewOnline(w Worker, c *config.VSwitch) (o *Online) {
	max := 128
	o = &Online{
		max:      max,
		lines:    make(map[string]*models.Line, max),
		lineList: list.New(),
		worker:   w,
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

	libol.Debug("Online.AddLine %s", line)
	if o.lineList.Len() > o.max {
		if e := o.lineList.Front(); e != nil {
			lastLine := e.Value.(*models.Line)

			o.lineList.Remove(e)
			delete(o.lines, lastLine.String())
			service.Online.Del(lastLine.String())
		}
	}

	libol.Debug("Online.AddLine %d", o.lineList.Len())
	find, ok := o.lines[line.String()]
	if !ok {
		o.lineList.PushBack(line)
		o.lines[line.String()] = line
		service.Online.Add(line)
	} else if find != nil {
		find.HitTime = time.Now().Unix()
	}
}
