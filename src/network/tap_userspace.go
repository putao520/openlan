package network

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"sync"
)

const (
	UsClose = uint(0x02)
	UsUp    = uint(0x04)
)

type UserSpaceTap struct {
	lock       sync.Mutex
	writeQueue chan []byte
	readQueue  chan []byte
	bridge     Bridger
	tenant     string
	flags      uint
	config     TapConfig
	name       string
	ifMtu      int
}

func NewUserSpaceTap(tenant string, c TapConfig) (*UserSpaceTap, error) {
	if c.Name == "" {
		c.Name = Tapers.GenName()
	}
	tap := &UserSpaceTap{
		tenant: tenant,
		name:   c.Name,
		ifMtu:  1514,
	}
	Tapers.Add(tap)

	return tap, nil
}

func (t *UserSpaceTap) Tenant() string {
	return t.tenant
}

func (t *UserSpaceTap) IsTun() bool {
	return t.config.Type == TUN
}

func (t *UserSpaceTap) IsTap() bool {
	return t.config.Type == TAP
}

func (t *UserSpaceTap) Name() string {
	return t.name
}

func (t *UserSpaceTap) hasFlags(flags uint) bool {
	return t.flags&flags == flags
}

func (t *UserSpaceTap) setFlags(flags uint) {
	t.flags |= flags
}

func (t *UserSpaceTap) clearFlags(flags uint) {
	t.flags &= ^flags
}

func (t *UserSpaceTap) Read(p []byte) (n int, err error) {
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	result := <-t.readQueue
	return copy(p, result), nil
}

func (t *UserSpaceTap) InRead(p []byte) (n int, err error) {
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("UserSpaceTap.InRead: %s % x", t, p[:20])
	}
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	t.readQueue <- p
	return len(p), nil
}

func (t *UserSpaceTap) Write(p []byte) (n int, err error) {
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("UserSpaceTap.Write: %s % x", t, p[:20])
	}
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()

	t.writeQueue <- p
	return len(p), nil
}

func (t *UserSpaceTap) OutWrite() ([]byte, error) {
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return nil, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	return <-t.writeQueue, nil
}

func (t *UserSpaceTap) Deliver() {
	for {
		data, err := t.OutWrite()
		if err != nil || data == nil {
			break
		}
		if libol.HasLog(libol.DEBUG) {
			libol.Debug("UserSpaceTap.Deliver: %s % x", t, data[:20])
		}
		if t.bridge == nil {
			continue
		}
		m := &Framer{Data: data, Source: t}
		_ = t.bridge.Input(m)
	}
}

func (t *UserSpaceTap) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.hasFlags(UsClose) {
		return nil
	}
	Tapers.Del(t.name)
	if t.bridge != nil {
		_ = t.bridge.DelSlave(t.name)
		t.bridge = nil
	}
	t.setFlags(UsClose)
	t.clearFlags(^UsUp)
	return nil
}

func (t *UserSpaceTap) Slave(bridge Bridger) {
	if t.bridge == nil {
		t.bridge = bridge
	}
}

func (t *UserSpaceTap) Up() {
	t.lock.Lock()
	t.writeQueue = make(chan []byte, 1024*32)
	t.readQueue = make(chan []byte, 1024*16)
	t.setFlags(UsUp)
	t.lock.Unlock()
	libol.Go(t.Deliver)
}

func (t *UserSpaceTap) Down() {
	t.lock.Lock()
	t.clearFlags(UsUp)
	t.writeQueue = nil
	t.readQueue = nil
	t.lock.Unlock()
}

func (t *UserSpaceTap) String() string {
	return t.name
}

func (t *UserSpaceTap) Mtu() int {
	return t.ifMtu
}

func (t *UserSpaceTap) SetMtu(mtu int) {
	t.ifMtu = mtu
}
