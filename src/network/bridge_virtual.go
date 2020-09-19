package network

import (
	"errors"
	"github.com/danieldin95/openlan-go/src/libol"
	"net"
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
	ports    map[string]Taper
	learners map[string]*Learner
	done     chan bool
	ticker   *time.Ticker
	timeout  int
	address  string
	kernel   Taper
	out      *libol.SubLogger
}

func NewVirtualBridge(name string, mtu int) *VirtualBridge {
	b := &VirtualBridge{
		name:     name,
		ifMtu:    mtu,
		ports:    make(map[string]Taper, 1024),
		learners: make(map[string]*Learner, 1024),
		done:     make(chan bool),
		ticker:   time.NewTicker(5 * time.Second),
		timeout:  5 * 60,
		out:      libol.NewSubLogger(name),
	}
	Bridges.Add(b)
	return b
}

func (b *VirtualBridge) Open(addr string) {
	b.out.Info("VirtualBridge.Open %s", addr)

	libol.Go(b.Start)
	if tap, err := NewKernelTap("", TapConfig{Type: TAP}); err != nil {
		b.out.Error("VirtualBridge.Open new kernel %s", err)
	} else {
		out, err := libol.IpLinkUp(tap.Name())
		if err != nil {
			b.out.Error("VirtualBridge.Open IpAddr %s:%s", err, out)
		}
		b.kernel = tap
		b.out.Info("VirtualBridge.Open %s", tap.Name())
		_ = b.AddSlave(tap.name)
	}
	if addr != "" && b.kernel != nil {
		b.address = addr
		if out, err := libol.IpAddrAdd(b.kernel.Name(), b.address); err != nil {
			b.out.Error("VirtualBridge.Open IpAddr %s:%s", err, out)
		}
	}
}

func (b *VirtualBridge) Kernel() string {
	if b.kernel == nil {
		return ""
	}
	return b.kernel.Name()
}

func (b *VirtualBridge) Close() error {
	if b.kernel != nil {
		if b.address != "" {
			out, err := libol.IpAddrDel(b.kernel.Name(), b.address)
			if err != nil {
				b.out.Error("VirtualBridge.Close: IpAddr %s:%s", err, out)
			}
		}
		b.kernel.Close()
	}
	b.ticker.Stop()
	b.done <- true
	return nil
}

func (b *VirtualBridge) AddSlave(name string) error {
	tap := Taps.Get(name)
	if tap == nil {
		return libol.NewErr("%s notFound", name)
	}
	tap.SetMtu(b.ifMtu) // consistent mtu value.
	b.lock.Lock()
	b.ports[name] = tap
	b.lock.Unlock()
	b.out.Info("VirtualBridge.AddSlave: %s", name)
	libol.Go(func() {
		for {
			data := make([]byte, b.ifMtu)
			n, err := tap.Recv(data)
			if err != nil || n == 0 {
				break
			}
			if libol.HasLog(libol.DEBUG) {
				libol.Debug("VirtualBridge.KernelTap: %s % x", tap.Name(), data[:20])
			}
			m := &Framer{Data: data[:n], Source: tap}
			_ = b.Input(m)
		}
	})
	return nil
}

func (b *VirtualBridge) DelSlave(name string) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if _, ok := b.ports[name]; ok {
		delete(b.ports, name)
	}
	b.out.Info("VirtualBridge.DelSlave: %s", name)
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
	if err := b.UniCast(m); err != nil {
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
	b.out.Debug("VirtualBridge.Expire delete %d", len(deletes))
	//execute delete.
	b.lock.Lock()
	for _, d := range deletes {
		if _, ok := b.learners[d]; ok {
			delete(b.learners, d)
			b.out.Info("VirtualBridge.Expire: delete %s", d)
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
				b.out.Log("VirtualBridge.Start: Tick at %s", t)
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
	if b.out.Has(libol.DEBUG) {
		b.out.Debug("VirtualBridge.Output: % x", m.Data[:20])
	}
	if dev := m.Output; dev != nil {
		_, err = dev.Send(m.Data)
	}
	return err
}

func (b *VirtualBridge) Eth2Str(addr []byte) string {
	if len(addr) < 6 {
		return ""
	}
	return net.HardwareAddr(addr).String()
}

func (b *VirtualBridge) Learn(m *Framer) {
	source := m.Data[6:12]
	if source[0]&0x01 == 0x01 {
		return
	}
	index := b.Eth2Str(source)
	if l := b.GetLearn(index); l != nil {
		b.UpdateLearn(index)
		return
	}
	learn := &Learner{
		Device:  m.Source,
		Uptime:  time.Now().Unix(),
		NewTime: time.Now().Unix(),
	}
	learn.Dest = make([]byte, 6)
	copy(learn.Dest, source)
	b.out.Info("VirtualBridge.Learn: %s on %s", index, m.Source)
	b.AddLearn(index, learn)
}

func (b *VirtualBridge) GetLearn(d string) *Learner {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if l, ok := b.learners[d]; ok {
		return l
	}
	return nil
}

func (b *VirtualBridge) AddLearn(d string, l *Learner) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.learners[d] = l
}

func (b *VirtualBridge) UpdateLearn(d string) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if l, ok := b.learners[d]; ok {
		l.Uptime = time.Now().Unix()
	}
}

func (b *VirtualBridge) Flood(m *Framer) error {
	data := m.Data
	src := m.Source
	if b.out.Has(libol.DEBUG) {
		b.out.Debug("VirtualBridge.Flood: % x", data[:20])
	}
	outs := make([]Taper, 0, 32)
	b.lock.RLock()
	for _, port := range b.ports {
		if src == port {
			continue
		}
		outs = append(outs, src)
	}
	b.lock.RUnlock()
	for _, port := range outs {
		if src == port {
			continue
		}
		if _, err := port.Send(data); err != nil {
			b.out.Error("VirtualBridge.Flood: %s %s", port, err)
		}
	}
	return nil
}

func (b *VirtualBridge) UniCast(m *Framer) error {
	data := m.Data
	src := m.Source
	index := b.Eth2Str(data[:6])
	learn := b.GetLearn(index)
	if learn == nil {
		return errors.New(index + " notFound")
	}
	dst := learn.Device
	if dst != src {
		if _, err := dst.Send(data); err != nil {
			b.out.Warn("VirtualBridge.UniCast: %s %s", dst, err)
		}
	}
	if b.out.Has(libol.DEBUG) {
		b.out.Debug("VirtualBridge.UniCast: %s to %s % x", src, dst, data[:20])
	}
	return nil
}

func (b *VirtualBridge) Mtu() int {
	return b.ifMtu
}

func (b *VirtualBridge) Stp(enable bool) error {
	return libol.NewErr("operation notSupport")
}

func (b *VirtualBridge) Delay(value int) error {
	return libol.NewErr("operation notSupport")
}
