package app

import (
	"container/list"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/switch/storage"
	"sync"
	"time"
)

type Online struct {
	lock     sync.RWMutex
	max      int
	lines    map[string]*models.Line
	lineList *list.List
	master   Master
}

func NewOnline(m Master, c config.Switch) *Online {
	max := 64
	return &Online{
		max:      max,
		lines:    make(map[string]*models.Line, max),
		lineList: list.New(),
		master:   m,
	}
}

func (o *Online) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	if frame.IsControl() {
		return nil
	}
	if libol.HasLog(libol.LOG) {
		libol.Log("Online.OnFrame %s.", frame)
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

func (o *Online) pop() {
	if o.lineList.Len() >= o.max {
		e := o.lineList.Front()
		if e == nil {
			return
		}
		lastLine, ok := e.Value.(*models.Line)
		if ok {
			o.lineList.Remove(e)
			storage.Online.Del(lastLine.String())
			delete(o.lines, lastLine.String())
		}
	}
}

func (o *Online) AddLine(line *models.Line) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if libol.HasLog(libol.LOG) {
		libol.Log("Online.AddLine %s", line)
		libol.Log("Online.AddLine %d", o.lineList.Len())
	}
	find, ok := o.lines[line.String()]
	if !ok {
		o.pop()
		o.lineList.PushBack(line)
		o.lines[line.String()] = line
		storage.Online.Add(line)
	} else if find != nil {
		find.HitTime = time.Now().Unix()
		storage.Online.Update(find)
	}
}
