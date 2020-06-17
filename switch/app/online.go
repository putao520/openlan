package app

import (
	"container/list"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/switch/storage"
	"sync"
	"time"
)

type Online struct {
	max      int
	lock     sync.RWMutex
	lines    map[string]*models.Line
	lineList *list.List
	master   Master
}

func NewOnline(m Master, c config.Switch) (o *Online) {
	max := 64
	o = &Online{
		max:      max,
		lines:    make(map[string]*models.Line, max),
		lineList: list.New(),
		master:   m,
	}
	return
}

func (o *Online) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	libol.Log("Online.OnFrame %s.", frame)
	if frame.IsControl() {
		return nil
	}
	proto, err := frame.Proto()
	if err != nil {
		libol.Warn("Online.OnFrame %s", err)
		return err
	}
	if proto.Ip4 != nil {
		ip := proto.Ip4
		line := models.NewLine(libol.EthIp4)
		line.IpSource = ip.Source
		line.IpDest = ip.Destination
		line.IpProtocol = ip.Protocol
		if proto.Tcp != nil {
			tcp := proto.Tcp
			line.PortDest = tcp.Destination
			line.PortSource = tcp.Source
		} else if proto.Udp != nil {
			udp := proto.Udp
			line.PortDest = udp.Destination
			line.PortSource = udp.Source
		} else {
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

	libol.Log("Online.AddLine %s", line)
	libol.Log("Online.AddLine %d", o.lineList.Len())
	find, ok := o.lines[line.String()]
	if !ok {
		if o.lineList.Len() >= o.max {
			if e := o.lineList.Front(); e != nil {
				lastLine := e.Value.(*models.Line)
				o.lineList.Remove(e)
				delete(o.lines, lastLine.String())
				storage.Online.Del(lastLine.String())
			}
		}
		o.lineList.PushBack(line)
		o.lines[line.String()] = line
		storage.Online.Add(line)
	} else if find != nil {
		find.HitTime = time.Now().Unix()
		storage.Online.Update(find)
	}
}
