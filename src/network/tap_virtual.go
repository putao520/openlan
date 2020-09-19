package network

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"sync"
)

const (
	UsClose = uint(0x02)
	UsUp    = uint(0x04)
)

type VirtualTap struct {
	lock    sync.Mutex
	kernel  chan []byte
	virtual chan []byte
	master  Bridger
	tenant  string
	flags   uint
	config  TapConfig
	name    string
	ifMtu   int
}

func NewVirtualTap(tenant string, c TapConfig) (*VirtualTap, error) {
	if c.Name == "" {
		c.Name = Taps.GenName()
	}
	tap := &VirtualTap{
		tenant: tenant,
		name:   c.Name,
		ifMtu:  1514,
		config: c,
	}
	Taps.Add(tap)
	return tap, nil
}

func (t *VirtualTap) Type() string {
	return "virtual"
}

func (t *VirtualTap) Tenant() string {
	return t.tenant
}

func (t *VirtualTap) IsTun() bool {
	return t.config.Type == TUN
}

func (t *VirtualTap) Name() string {
	return t.name
}

func (t *VirtualTap) hasFlags(flags uint) bool {
	return t.flags&flags == flags
}

func (t *VirtualTap) setFlags(flags uint) {
	t.flags |= flags
}

func (t *VirtualTap) clearFlags(flags uint) {
	t.flags &= ^flags
}

func (t *VirtualTap) Write(p []byte) (int, error) {
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("VirtualTap.Write: %s % x", t, p[:20])
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

func (t *VirtualTap) Read(p []byte) (int, error) {
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	data := <-t.kernel
	return copy(p, data), nil
}

func (t *VirtualTap) Recv(p []byte) (int, error) {
	t.lock.Lock()
	if !t.hasFlags(UsUp) {
		t.lock.Unlock()
		return 0, libol.NewErr("notUp")
	}
	t.lock.Unlock()
	data := <-t.virtual
	return copy(p, data), nil
}

func (t *VirtualTap) Send(p []byte) (int, error) {
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("VirtualTap.Send: %s % x", t, p[:20])
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

func (t *VirtualTap) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.hasFlags(UsClose) {
		return nil
	}
	Taps.Del(t.name)
	if t.master != nil {
		_ = t.master.DelSlave(t.name)
		t.master = nil
	}
	t.setFlags(UsClose)
	t.clearFlags(UsUp)
	return nil
}

func (t *VirtualTap) Master(dev Bridger) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.master == nil {
		t.master = dev
	} else {
		libol.Warn("VirtualTap.Master already for %s", t.master)
	}
}

func (t *VirtualTap) Up() {
	t.lock.Lock()
	t.kernel = make(chan []byte, t.config.SendBuf)
	t.virtual = make(chan []byte, t.config.WriteBuf)
	t.setFlags(UsUp)
	t.lock.Unlock()
}

func (t *VirtualTap) Down() {
	t.lock.Lock()
	t.clearFlags(UsUp)
	close(t.kernel)
	t.kernel = nil
	close(t.virtual)
	t.virtual = nil
	t.lock.Unlock()
}

func (t *VirtualTap) String() string {
	return t.name
}

func (t *VirtualTap) Mtu() int {
	return t.ifMtu
}

func (t *VirtualTap) SetMtu(mtu int) {
	t.ifMtu = mtu
}
