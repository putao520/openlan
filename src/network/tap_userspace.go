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
	lock    sync.Mutex
	kernel  chan []byte
	virtual chan []byte
	bridge  Bridger
	tenant  string
	flags   uint
	config  TapConfig
	name    string
	ifMtu   int
}

func NewUserSpaceTap(tenant string, c TapConfig) (*UserSpaceTap, error) {
	if c.Name == "" {
		c.Name = Tapers.GenName()
	}
	tap := &UserSpaceTap{
		tenant: tenant,
		name:   c.Name,
		ifMtu:  1514,
		config: c,
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

func (t *UserSpaceTap) Write(p []byte) (int, error) {
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("UserSpaceTap.Write: %s % x", t, p[:20])
	}
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	t.virtual <- p
	return len(p), nil
}

func (t *UserSpaceTap) Read(p []byte) (int, error) {
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	data := <-t.kernel
	return copy(p, data), nil
}

func (t *UserSpaceTap) Recv(p []byte) (int, error) {
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	data := <-t.virtual
	return copy(p, data), nil
}

func (t *UserSpaceTap) Send(p []byte) (int, error) {
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("UserSpaceTap.Send: %s % x", t, p[:20])
	}
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	t.kernel <- p
	return len(p), nil
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
	t.kernel = make(chan []byte, 1024*32)
	t.virtual = make(chan []byte, 1024*16)
	t.setFlags(UsUp)
	t.lock.Unlock()
}

func (t *UserSpaceTap) Down() {
	t.lock.Lock()
	t.clearFlags(UsUp)
	t.kernel = nil
	t.virtual = nil
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
