package app

import (
	"github.com/danieldin95/openlan-go/network"
)

type Master interface {
	ReadTap(dev network.Taper, readAt func(p []byte) error)
	NewTap(tenant string) (network.Taper, error)
	UUID() string
}
