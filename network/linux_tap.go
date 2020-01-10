package network

import (
	"github.com/songgao/water"
)

type LinuxTap struct {
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
		isTap:  device.IsTAP(),
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
	return t.device.Read(p)
}

func (t *LinuxTap) InRead(p []byte) (n int, err error) {
	//TODO
	return 0, nil
}

func (t *LinuxTap) Write(p []byte) (n int, err error) {
	return t.device.Write(p)
}

func (t *LinuxTap) Close() error {
	Tapers.Del(t.name)
	if t.bridge != nil {
		t.bridge.DelSlave(t)
		t.bridge = nil
	}
	return t.device.Close()
}

func (t *LinuxTap) Slave(bridge Bridger) {
	if t.bridge == nil {
		t.bridge = bridge
	}
}

func (t LinuxTap) Up() {
	//TODO
}

func (t *LinuxTap) String() string {
	return t.name
}
