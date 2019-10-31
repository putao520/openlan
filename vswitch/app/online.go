package app

import (
	"fmt"
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/models"
	"github.com/lightstar-dev/openlan-go/vswitch/api"
	"strings"
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

func (o *Online) AddLine(line *models.Line) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if _, ok := o.lines[line.String()]; !ok {
		libol.Info("Online.AddLine %s", line)
		o.lines[line.String()] = line
		o.PubLine(line, true)
	}
}

func (o *Online) OnClientClose(client *libol.TcpClient) {
	//TODO
}

func (o *Online) RedisId() string {
	wid := strings.Replace(o.worker.GetId(), ":", "/", -1)
	return fmt.Sprintf("%s:online", wid)
}

func (o *Online) PubLine(l *models.Line, isAdd bool) {
	lid := strings.Replace(l.String(), ":", "/", -1)
	key := fmt.Sprintf("%s:%s", o.RedisId(), lid)
	value := map[string]interface{}{
		"ethernet":  fmt.Sprintf("0x%04x", l.EthType),
		"source":  l.IpSource.String(),
		"destination":  l.IPDest.String(),
		"protocol": fmt.Sprintf("0x%02x", l.IpProtocol),
		"port": fmt.Sprintf("%d", l.PortDest),
	}

	if r := o.worker.GetRedis(); r != nil {
		if err := r.HMSet(key, value); err != nil {
			libol.Error("Online.PubLine HMSet %s", err)
		}
	}
}