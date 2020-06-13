package network

import (
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"sync"
	"time"
)

type Learner struct {
	Dest    []byte
	Device  Taper
	Uptime  int64
	NewTime int64
}

type VirtualBridge struct {
	ifMtu    int
	name     string
	lock     sync.RWMutex
	devices  map[string]Taper
	learners map[string]*Learner
	done     chan bool
	ticker   *time.Ticker
	timeout  int
	address  string
	device   Taper
}

func NewVirtualBridge(name string, mtu int) *VirtualBridge {
	b := &VirtualBridge{
		name:     name,
		ifMtu:    mtu,
		devices:  make(map[string]Taper, 1024),
		learners: make(map[string]*Learner, 1024),
		done:     make(chan bool),
		ticker:   time.NewTicker(5 * time.Second),
		timeout:  5 * 60,
	}
	return b
}

func (b *VirtualBridge) Open(addr string) {
	libol.Info("VirtualBridge.Open %s", addr)
	if addr != "" {
		tap, err := NewKernelTap("default", TapConfig{Type: TAP})
		if err != nil {
			libol.Error("VirtualBridge.Open new kernel %s", err)
		} else {
			out, err := libol.IpLinkUp(tap.Name())
			if err != nil {
				libol.Error("VirtualBridge.Open.IpAddr %s:%s", err, out)
			}
			b.address = addr
			b.device = tap
			out, err = libol.IpAddrAdd(b.device.Name(), b.address)
			if err != nil {
				libol.Error("VirtualBridge.Open.IpAddr %s:%s", err, out)
			}
			libol.Info("VirtualBridge.Open %s", tap.Name())
		}
	} else {
		libol.Warn("VirtualBridge.Open: not support address")
	}
	libol.Go(b.Start)
}

func (b *VirtualBridge) Close() error {
	if b.device != nil {
		out, err := libol.IpAddrDel(b.device.Name(), b.address)
		if err != nil {
			libol.Error("VirtualBridge.Close.IpAddr %s:%s", err, out)
		}
	}
	b.ticker.Stop()
	b.done <- true
	return nil
}

func (b *VirtualBridge) AddSlave(dev Taper) error {
	dev.Slave(b)

	b.lock.Lock()
	defer b.lock.Unlock()
	b.devices[dev.Name()] = dev

	libol.Info("VirtualBridge.AddSlave: %s %s", dev.Name(), b.name)

	return nil
}

func (b *VirtualBridge) DelSlave(dev Taper) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if _, ok := b.devices[dev.Name()]; ok {
		delete(b.devices, dev.Name())
	}

	libol.Info("VirtualBridge.DelSlave: %s %s", dev.Name(), b.name)

	return nil
}

func (b *VirtualBridge) Type() string {
	return "virtual"
}

func (b *VirtualBridge) Name() string {
	return b.name
}

func (b *VirtualBridge) SetName(value string) {
	b.name = value
}

func (b *VirtualBridge) SetTimeout(value int) {
	b.timeout = value
}

func (b *VirtualBridge) Forward(m *Framer) error {
	if is := b.Unicast(m); !is {
		_ = b.Flood(m)
	}
	return nil
}

func (b *VirtualBridge) Expire() error {
	deletes := make([]string, 0, 1024)

	//collect need deleted.
	b.lock.RLock()
	for index, learn := range b.learners {
		now := time.Now().Unix()
		if now-learn.Uptime > int64(b.timeout) {
			deletes = append(deletes, index)
		}
	}
	b.lock.RUnlock()

	libol.Debug("VirtualBridge.Expire delete %d", len(deletes))
	//execute delete.
	b.lock.Lock()
	for _, d := range deletes {
		if _, ok := b.learners[d]; ok {
			delete(b.learners, d)
			libol.Info("VirtualBridge.Expire: delete %s", d)
		}
	}
	b.lock.Unlock()

	return nil
}

func (b *VirtualBridge) Start() {
	libol.Go(func() {
		for {
			select {
			case <-b.done:
				return
			case t := <-b.ticker.C:
				libol.Debug("VirtualBridge.Expire Tick at %s", t)
				_ = b.Expire()
			}
		}
	})
}

func (b *VirtualBridge) Input(m *Framer) error {
	b.Learn(m)
	return b.Forward(m)
}

func (b *VirtualBridge) Output(m *Framer) error {
	var err error

	libol.Debug("VirtualBridge.Output: % x", m.Data[:20])
	if dev := m.Output; dev != nil {
		_, err = dev.InRead(m.Data)
	}

	return err
}

func (b *VirtualBridge) Eth2Str(addr []byte) string {
	if len(addr) < 6 {
		return ""
	}
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x",
		addr[0], addr[1], addr[2], addr[3], addr[4], addr[5])
}

func (b *VirtualBridge) Learn(m *Framer) {
	source := m.Data[6:12]
	if source[0]&0x01 == 0x01 {
		return
	}

	index := b.Eth2Str(source)
	if l := b.FindDest(index); l != nil {
		b.UpdateDest(index)
		return
	}

	learn := &Learner{
		Device:  m.Source,
		Uptime:  time.Now().Unix(),
		NewTime: time.Now().Unix(),
	}
	learn.Dest = make([]byte, 6)
	copy(learn.Dest, source)

	libol.Info("VirtualBridge.Learn: %s on %s", index, m.Source)
	b.AddDest(index, learn)
}

func (b *VirtualBridge) FindDest(d string) *Learner {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if l, ok := b.learners[d]; ok {
		return l
	}
	return nil
}

func (b *VirtualBridge) AddDest(d string, l *Learner) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.learners[d] = l
}

func (b *VirtualBridge) UpdateDest(d string) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if l, ok := b.learners[d]; ok {
		l.Uptime = time.Now().Unix()
	}
}

func (b *VirtualBridge) Flood(m *Framer) error {
	var err error

	data := m.Data
	src := m.Source
	libol.Debug("VirtualBridge.Flood: % x", data[:20])
	for _, dst := range b.devices {
		if src == dst {
			continue
		}
		_, err = dst.InRead(data)
	}
	return err
}

func (b *VirtualBridge) Unicast(m *Framer) bool {
	data := m.Data
	src := m.Source
	index := b.Eth2Str(data[:6])

	if l := b.FindDest(index); l != nil {
		dst := l.Device
		if dst != src {
			if _, err := dst.InRead(data); err != nil {
				libol.Debug("VirtualBridge.Unicast: %s %s", dst, err)
			}
		}
		libol.Debug("VirtualBridge.Unicast: %s to %s % x", src, dst, data[:20])
		return true
	}

	return false
}

func (b *VirtualBridge) Mtu() int {
	return b.ifMtu
}
