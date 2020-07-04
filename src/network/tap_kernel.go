package network

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/songgao/water"
)

type KernelTap struct {
	lock   libol.Locker
	device *water.Interface
	bridge Bridger
	tenant string
	name   string
	config TapConfig
	ifMtu  int
}

func NewKernelTap(tenant string, c TapConfig) (*KernelTap, error) {
	device, err := WaterNew(c)
	if err != nil {
		return nil, err
	}
	tap := &KernelTap{
		tenant: tenant,
		device: device,
		name:   device.Name(),
		config: c,
		ifMtu:  1514,
	}

	Tapers.Add(tap)

	return tap, nil
}

func (t *KernelTap) Tenant() string {
	return t.tenant
}

func (t *KernelTap) IsTun() bool {
	return t.config.Type == TUN
}

func (t *KernelTap) IsTap() bool {
	return t.config.Type == TAP
}

func (t *KernelTap) Name() string {
	return t.name
}

func (t *KernelTap) Read(p []byte) (n int, err error) {
	t.lock.Lock()
	if t.device == nil {
		t.lock.Unlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.Unlock()

	return t.device.Read(p)
}

func (t *KernelTap) InRead(p []byte) (n int, err error) {
	t.lock.Lock()
	if t.device == nil {
		t.lock.Unlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.Unlock()

	return 0, nil
}

func (t *KernelTap) Write(p []byte) (n int, err error) {
	t.lock.Lock()
	if t.device == nil {
		t.lock.Unlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.Unlock()

	return t.device.Write(p)
}

func (t *KernelTap) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.device == nil {
		return nil
	}

	Tapers.Del(t.name)
	if t.bridge != nil {
		_ = t.bridge.DelSlave(t)
		t.bridge = nil
	}
	err := t.device.Close()
	t.device = nil

	return err
}

func (t *KernelTap) Slave(bridge Bridger) {
	if t.bridge == nil {
		t.bridge = bridge
	}
}

func (t *KernelTap) Up() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.device == nil {
		device, err := WaterNew(t.config)
		if err != nil {
			libol.Error("KernelTap.Up %s", err)
			return
		}
		t.device = device
		Tapers.Add(t)
	}
}

func (t *KernelTap) String() string {
	return t.name
}

func (t *KernelTap) Mtu() int {
	return t.ifMtu
}

func (t *KernelTap) SetMtu(mtu int) {
	t.ifMtu = mtu
}
