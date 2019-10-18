package vswitch

import (
	"fmt"
	"github.com/lightstar-dev/openlan-go/libol"
	"net"
	"sync"
)

type Line struct {
	EthType uint16
	IpAddr  net.IP
	IpProto uint8
	L4Port  uint16
}

func NewLine(t uint16) *Line {
	l := &Line{
		EthType: t,
		IpAddr:  nil,
		IpProto: 0,
		L4Port:  0,
	}
	return l
}

func (l *Line) String() string {
	return fmt.Sprintf("%d:%s:%d:%d", l.EthType, l.IpAddr, l.IpProto, l.L4Port)
}

type Online struct {
	lock   sync.RWMutex
	lines  map[string]*Line
	worker WorkerApi
}

func NewOnline(api WorkerApi, c *Config) (o *Online) {
	o = &Online{
		lines:  make(map[string]*Line, 1024*4),
		worker: api,
	}
	return
}

func (o *Online) GetOnline(name string) *Line {
	o.lock.RLock()
	defer o.lock.RUnlock()

	if n, ok := o.lines[name]; ok {
		return n
	}

	return nil
}

func (o *Online) ListLine() <-chan *Line {
	c := make(chan *Line, 1024)

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
	libol.Debug("Onlino.OnFrame % x.", data)

	if libol.IsInst(data) {
		return nil
	}

	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("Onlino.OnFrame %s", err)
		return err
	}

	data = data[eth.Len:]
	if eth.IsIP4() {
		ip, err := libol.NewIpv4FromFrame(data)
		if err != nil {
			libol.Warn("Onlino.OnFrame %s", err)
			return err
		}
		data = data[ip.Len:]

		line := NewLine(eth.Type)
		line.IpAddr = ip.Destination
		line.IpProto = ip.Protocol

		switch ip.Protocol {
		case libol.IPPROTO_ICMP:
			//TODO
		case libol.IPPROTO_TCP:
			tcp, err := libol.NewTcpFromFrame(data)
			if err != nil {
				libol.Warn("Onlino.OnFrame %s", err)
			}
			line.L4Port = tcp.Destination
		case libol.IPPROTO_UDP:
			udp, err := libol.NewUdpFromFrame(data)
			if err != nil {
				libol.Warn("Onlino.OnFrame %s", err)
			}
			line.L4Port = udp.Destination
		default:
			line.L4Port = 0
		}

		o.AddLine(line)
	}

	return nil
}

func (o *Online) AddLine(line *Line) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if _, ok := o.lines[line.String()]; !ok {
		libol.Info("OnLino.AddLine %s", line)
		o.lines[line.String()] = line
	}
}

func (o *Online) OnClientClose(client *libol.TcpClient) {
	//TODO
}
