package models

import "github.com/songgao/water"

type TapDevice struct {
	*water.Interface
}

func NewTapDevice(dev *water.Interface) *TapDevice {
	t := &TapDevice{
		Interface: dev,
	}

	return t
}
