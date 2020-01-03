package network

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/golang-collections/go-datastructures/queue"
	"sync"
)

type VirBridge struct {
	mtu     int
	name    string
	inQ     *queue.RingBuffer
	lock    sync.RWMutex
	devices map[string]Taper
	dests   map[string]Taper
}

func NewVirBridge(name string, mtu int) *VirBridge {
	b := &VirBridge{
		name: name,
		mtu:  mtu,
		inQ:  queue.NewRingBuffer(1024*20),
		devices: make(map[string]Taper, 1024),
		dests:   make(map[string]Taper, 1024),
	}
	return b
}

func (b *VirBridge) Open(addr string) {
	libol.Info("VirBridge.Open: not support address")
	go b.Start()
}

func (b *VirBridge) Close() error {
	b.inQ.Dispose()
	return nil
}

func (b *VirBridge) AddSlave(dev Taper) error {
	dev.Slave(b)

	b.lock.Lock()
	b.devices[dev.Name()] = dev
	b.lock.Unlock()

	libol.Info("VirBridge.AddSlave: %s %s", dev.Name(), b.name)

	return nil
}

func (b *VirBridge) DelSlave(dev Taper) error {
	b.lock.Lock()
	if _, ok := b.devices[dev.Name()]; ok {
		delete(b.devices, dev.Name())
	}
	b.lock.Unlock()

	libol.Info("VirBridge.DelSlave: %s %s", dev.Name(), b.name)

	return nil
}

func (b *VirBridge) Name() string {
	return b.name
}

func (b *VirBridge) SetName(value string) {
	b.name = value
}

func (b *VirBridge) Start() {
	for {
		result, err := b.inQ.Get()
		if err != nil {
			return
		}
		b.Flood(result.(*Framer))
	}
}

func (b *VirBridge) Input(m *Framer) error {
	return b.inQ.Put(m)
}

func (b *VirBridge) Output(m *Framer) error {
	var err error

	libol.Debug("VirBridge.Output: % x", m.Data[:20])
	if dev := m.Output; dev != nil {
		_, err = dev.InRead(m.Data)
	}

	return err
}

func (b *VirBridge) FindDest(d string) Taper {
	return nil
}

func (b *VirBridge) AddDest(d string, t Taper) {
}

func (b *VirBridge) Update(d string, t Taper) {

}

func (b *VirBridge) Flood(m *Framer) error {
	var err error

	libol.Debug("VirBridge.Flood: % x", m.Data[:20])
	for _, dev := range b.devices {
		if m.Source == dev {
			continue
		}

		_, err = dev.InRead(m.Data)
	}
	return err
}