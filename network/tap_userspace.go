package network

import (
	"github.com/danieldin95/openlan-go/libol"
	"sync"
)

type UserSpaceTap struct {
	lock   sync.RWMutex
	closed bool
	isTap  bool
	name   string
	writeQ chan []byte
	readQ  chan []byte
	bridge Bridger
	tenant string
	mtu    int
}

func NewUserSpaceTap(isTap bool, tenant, name string) (*UserSpaceTap, error) {
	if name == "" {
		name = Tapers.GenName()
	}
	tap := &UserSpaceTap{
		tenant: tenant,
		isTap:  isTap,
		name:   name,
		mtu:    1514,
	}
	Tapers.Add(tap)

	return tap, nil
}

func (t *UserSpaceTap) Tenant() string {
	return t.tenant
}

func (t *UserSpaceTap) IsTun() bool {
	return !t.isTap
}

func (t *UserSpaceTap) IsTap() bool {
	return t.isTap
}

func (t *UserSpaceTap) Name() string {
	return t.name
}

func (t *UserSpaceTap) Read(p []byte) (n int, err error) {
	t.lock.RLock()
	if t.closed || t.readQ == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Close")
	}
	t.lock.RUnlock()

	result := <-t.readQ
	return copy(p, result), nil
}

func (t *UserSpaceTap) InRead(p []byte) (n int, err error) {
	libol.Debug("UserSpaceTap.InRead: %s % x", t, p[:20])
	t.lock.RLock()
	if t.closed || t.readQ == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Close")
	}
	t.lock.RUnlock()

	t.readQ <- p
	return len(p), nil
}

func (t *UserSpaceTap) Write(p []byte) (n int, err error) {
	libol.Debug("UserSpaceTap.Write: %s % x", t, p[:20])
	t.lock.RLock()
	if t.closed || t.writeQ == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Close")
	}
	t.lock.RUnlock()

	t.writeQ <- p
	return len(p), nil
}

func (t *UserSpaceTap) OutWrite() ([]byte, error) {
	t.lock.RLock()
	if t.closed || t.writeQ == nil {
		t.lock.RUnlock()
		return nil, libol.NewErr("Close")
	}
	t.lock.RUnlock()

	return <-t.writeQ, nil
}

func (t *UserSpaceTap) Deliver() {
	for {
		data, err := t.OutWrite()
		if err != nil || data == nil {
			break
		}
		libol.Debug("UserSpaceTap.Deliver: %s % x", t, data[:20])
		if t.bridge == nil {
			continue
		}

		m := &Framer{Data: data, Source: t}
		t.bridge.Input(m)
	}
}

func (t *UserSpaceTap) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.closed {
		return nil
	}

	close(t.readQ)
	close(t.writeQ)
	Tapers.Del(t.name)
	if t.bridge != nil {
		t.bridge.DelSlave(t)
		t.bridge = nil
	}
	t.readQ = nil
	t.writeQ = nil
	t.closed = true

	return nil
}

func (t *UserSpaceTap) Slave(bridge Bridger) {
	if t.bridge == nil {
		t.bridge = bridge
	}
}

func (t *UserSpaceTap) Up() {
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

func (t *UserSpaceTap) String() string {
	return t.name
}

func (t *UserSpaceTap) Mtu() int {
	return t.mtu
}

func (t *UserSpaceTap) SetMtu(mtu int) {
	t.mtu = mtu
}
