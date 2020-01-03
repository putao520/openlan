package network

import (
	"github.com/songgao/water"
)

type LinTap struct {
	dev *water.Interface
}

func NewLinTap(isTap bool, name string) (*LinTap, error) {
	devType := water.DeviceType(water.TUN)
	if isTap {
		devType = water.TAP
	}

	dev, err := water.New(water.Config{DeviceType: devType})
	if err != nil {
		return nil, err
	}

	tap := &LinTap{
		dev: dev,
	}

	return tap, nil
}

func (t *LinTap) IsTUN() bool {
	return !t.dev.IsTUN()
}

func (t *LinTap) IsTAP() bool {
	return t.dev.IsTAP()
}

func (t *LinTap) Name() string {
	return t.dev.Name()
}

func (t *LinTap) Read(p []byte) (n int, err error) {
	return t.dev.Read(p)
}

func (t *LinTap) Write(p []byte) (n int, err error) {
	return t.dev.Write(p)
}

func (t *LinTap) Close() error {
	return t.Close()
}
