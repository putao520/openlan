package network

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/songgao/water"
	"sync"
)

type LinuxTap struct {
	lock   sync.RWMutex
	isTap  bool
	name   string
	device *water.Interface
	bridge Bridger
}

func NewLinuxTap(isTap bool, name string) (*LinuxTap, error) {
	deviceType := water.DeviceType(water.TUN)
	if isTap {
		deviceType = water.TAP
	}
	device, err := water.New(water.Config{DeviceType: deviceType})
	if err != nil {
		return nil, err
	}
	tap := &LinuxTap{
		device: device,
		name:   device.Name(),
		isTap:  isTap,
	}

	Tapers.Add(tap)

	return tap, nil
}

func (t *LinuxTap) IsTun() bool {
	return !t.isTap
}

func (t *LinuxTap) IsTap() bool {
	return t.isTap
}

func (t *LinuxTap) Name() string {
	return t.name
}

func (t *LinuxTap) Read(p []byte) (n int, err error) {
	t.lock.RLock()
	if t.device == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.RUnlock()

	return t.device.Read(p)
}

func (t *LinuxTap) InRead(p []byte) (n int, err error) {
	t.lock.RLock()
	if t.device == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.RUnlock()

	return 0, nil
}

func (t *LinuxTap) Write(p []byte) (n int, err error) {
	t.lock.RLock()
	if t.device == nil {
		t.lock.RUnlock()
		return 0, libol.NewErr("Closed")
	}
	t.lock.RUnlock()

	return t.device.Write(p)
}

func (t *LinuxTap) Close() error {
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

func (t *LinuxTap) Slave(bridge Bridger) {
	if t.bridge == nil {
		t.bridge = bridge
	}
}

func (t *LinuxTap) Up() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.device == nil {
		deviceType := water.DeviceType(water.TUN)
		if t.IsTap() {
			deviceType = water.TAP
		}
		device, err := water.New(water.Config{DeviceType: deviceType})
		if err != nil {
			libol.Error("LinuxTap.Up %s", err)
			return
		}
		t.device = device
		Tapers.Add(t)
	}
}

func (t *LinuxTap) String() string {
	return t.name
}
