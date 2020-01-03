package network

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/golang-collections/go-datastructures/queue"
	"sync"
)

type Framer struct {
	Data   []byte
	Device Taper
}

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
	}
	return b
}

func (b *VirBridge) Open(addr string) {
	b.inQ = queue.NewRingBuffer(1024*20)

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

		m := result.(Framer)
		b.Flood(m.Data)
	}
}

func (b *VirBridge) Input(p []byte, t Taper) error {
	m := &Framer{Data: p, Device: t}
	return b.inQ.Put(m)
}

func (b *VirBridge) Output(p []byte, t Taper) error {
	return nil
}

func (b *VirBridge) FindDest(dest string) Taper {
	return nil
}

func (b *VirBridge) AddDest(dest string, t Taper) {

}

func (b *VirBridge) Update(dest string, t Taper) {

}

func (b *VirBridge) Flood(p []byte) error {
	var err error

	for _, dev := range b.devices {
		_, err = dev.Write(p)
	}
	return err
}