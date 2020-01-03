package network

import (
	"github.com/songgao/water"
)

type LinTap struct {
	isTap bool
	name  string
	device   *water.Interface
	bridge    Bridger
}

func NewLinTap(isTap bool, name string) (*LinTap, error) {
	deviceType := water.DeviceType(water.TUN)
	if isTap {
		deviceType = water.TAP
	}
	device, err := water.New(water.Config{DeviceType: deviceType})
	if err != nil {
		return nil, err
	}
	tap := &LinTap{
		device: device,
		name: device.Name(),
		isTap: device.IsTAP(),
	}

	Tapers.Add(tap)

	return tap, nil
}

func (t *LinTap) IsTUN() bool {
	return !t.isTap
}

func (t *LinTap) IsTAP() bool {
	return t.isTap
}

func (t *LinTap) Name() string {
	return t.name
}

func (t *LinTap) Read(p []byte) (n int, err error) {
	return t.device.Read(p)
}

func (t *LinTap) Write(p []byte) (n int, err error) {
	return t.device.Write(p)
}

func (t *LinTap) Close() error {
	Tapers.Del(t.name)
	if t.bridge != nil {
		t.bridge.DelSlave(t)
		t.bridge = nil
	}
	return t.device.Close()
}

func (t *LinTap) Slave(bridge Bridger) {
	if t.bridge == nil {
		t.bridge = bridge
	}
}