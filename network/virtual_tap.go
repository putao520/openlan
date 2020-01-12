package network

import (
	"github.com/danieldin95/openlan-go/libol"
	"sync"
)

type VirtualTap struct {
	lock   sync.RWMutex
	closed bool
	isTap  bool
	name   string
	writeQ chan []byte
	readQ  chan []byte
	bridge Bridger
}

func NewVirtualTap(isTap bool, name string) (*VirtualTap, error) {
	if name == "" {
		name = Tapers.GenName()
	}
	tap := &VirtualTap{
		isTap: isTap,
		name:  name,
	}
	Tapers.Add(tap)

	return tap, nil
}

func (t *VirtualTap) IsTun() bool {
	return !t.isTap
}

func (t *VirtualTap) IsTap() bool {
	return t.isTap
}

func (t *VirtualTap) Name() string {
	return t.name
}

func (t *VirtualTap) Read(p []byte) (n int, err error) {
	t.lock.RLock()
	if t.closed || t.readQ == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Close")
	}
	t.lock.RUnlock()

	result := <-t.readQ
	return copy(p, result), nil
}

func (t *VirtualTap) InRead(p []byte) (n int, err error) {
	libol.Debug("VirtualTap.InRead: %s % x", t, p[:20])
	t.lock.RLock()
	if t.closed || t.readQ == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Close")
	}
	t.lock.RUnlock()

	t.readQ <- p
	return len(p), nil
}

func (t *VirtualTap) Write(p []byte) (n int, err error) {
	libol.Debug("VirtualTap.Write: %s % x", t, p[:20])
	t.lock.RLock()
	if t.closed || t.writeQ == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Close")
	}
	t.lock.RUnlock()

	t.writeQ <- p
	return len(p), nil
}

func (t *VirtualTap) OutWrite() ([]byte, error) {
	t.lock.RLock()
	if t.closed || t.writeQ == nil {
		t.lock.RUnlock()
		return nil, libol.NewErr("Close")
	}
	t.lock.RUnlock()

	return <-t.writeQ, nil
}

func (t *VirtualTap) Deliver() {
	for {
		data, err := t.OutWrite()
		if err != nil || data == nil {
			break
		}
		libol.Debug("VirtualTap.Deliver: %s % x", t, data[:20])
		if t.bridge == nil {
			continue
		}

		m := &Framer{Data: data, Source: t}
		t.bridge.Input(m)
	}
}

func (t *VirtualTap) Close() error {
	t.lock.Lock()
	if t.closed {
		t.lock.Unlock()
		return nil
	}
	t.closed = true
	t.lock.Unlock()

	close(t.readQ)
	close(t.writeQ)
	if t.bridge != nil {
		t.bridge.DelSlave(t)
		t.bridge = nil
	}
	t.readQ = nil
	t.writeQ = nil
	Tapers.Del(t.name)

	return nil
}

func (t *VirtualTap) Slave(bridge Bridger) {
	if t.bridge == nil {
		t.bridge = bridge
	}
}

func (t *VirtualTap) Up() {
	t.lock.Lock()
	if t.closed {
		Tapers.Add(t)
	}
	if t.writeQ == nil {
		t.writeQ = make(chan []byte, 1024*32)
	}
	if t.readQ == nil {
		t.readQ = make(chan []byte, 1024*16)
	}
	t.closed = false
	t.lock.Unlock()

	go t.Deliver()
}

func (t *VirtualTap) String() string {
	return t.name
}
