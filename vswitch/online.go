package vswitch

import (
	"fmt"
	"github.com/lightstar-dev/openlan-go/libol"
	"net"
	"sync"
)

type Line struct {
	EthType    uint16
	IpSource   net.IP
	IPDest     net.IP
	IpProtocol uint8
	PortDest   uint16
	PortSource uint16
}

func NewLine(t uint16) *Line {
	l := &Line{
		EthType: t,
		IpSource:  nil,
		IpProtocol: 0,
		PortDest:  0,
	}
	return l
}

func (l *Line) String() string {
	return fmt.Sprintf("%d:%s:%s:%d:%d",
		l.EthType, l.IpSource, l.IPDest, l.IpProtocol, l.PortDest)
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
	libol.Debug("Online.OnFrame % x.", data)

	if libol.IsInst(data) {
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

		line := NewLine(eth.Type)
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
		case libol.IPPROTO_UDP:
			udp, err := libol.NewUdpFromFrame(data)
			if err != nil {
				libol.Warn("Online.OnFrame %s", err)
			}
			line.PortDest = udp.Destination
		default:
			line.PortSource = 0
		}

		o.AddLine(line)
	}

	return nil
}

func (o *Online) AddLine(line *Line) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if _, ok := o.lines[line.String()]; !ok {
		libol.Info("Online.AddLine %s", line)
		o.lines[line.String()] = line
	}
}

func (o *Online) OnClientClose(client *libol.TcpClient) {
	//TODO
}
