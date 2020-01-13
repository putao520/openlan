package network

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/songgao/water"
	"sync"
)

type KernelTap struct {
	lock   sync.RWMutex
	isTap  bool
	name   string
	device *water.Interface
	bridge Bridger
}

func NewKernelTap(isTap bool, name string) (*KernelTap, error) {
	deviceType := water.DeviceType(water.TUN)
	if isTap {
		deviceType = water.TAP
	}
	device, err := water.New(water.Config{DeviceType: deviceType})
	if err != nil {
		return nil, err
	}
	tap := &KernelTap{
		device: device,
		name:   device.Name(),
		isTap:  isTap,
	}

	Tapers.Add(tap)

	return tap, nil
}

func (t *KernelTap) IsTun() bool {
	return !t.isTap
}

func (t *KernelTap) IsTap() bool {
	return t.isTap
}

func (t *KernelTap) Name() string {
	return t.name
}

func (t *KernelTap) Read(p []byte) (n int, err error) {
	t.lock.RLock()
	if t.device == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.RUnlock()

	return t.device.Read(p)
}

func (t *KernelTap) InRead(p []byte) (n int, err error) {
	t.lock.RLock()
	if t.device == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.RUnlock()

	return 0, nil
}

func (t *KernelTap) Write(p []byte) (n int, err error) {
	t.lock.RLock()
	if t.device == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.RUnlock()

	return t.device.Write(p)
}

func (t *KernelTap) Close() error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.device == nil {
		return nil
	}

	Tapers.Del(t.name)
	if t.bridge != nil {
		t.bridge.DelSlave(t)
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
		deviceType := water.DeviceType(water.TUN)
		if t.IsTap() {
			deviceType = water.TAP
		}
		device, err := water.New(water.Config{DeviceType: deviceType})
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
