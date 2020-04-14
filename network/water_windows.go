package network

import (
	"github.com/songgao/water"
)

func WaterNew(c TapConfig) (*water.Interface, error) {
	deviceType := water.DeviceType(water.TAP)
	if c.Type == TUN {
		deviceType = water.TUN
	}
	cfg := water.Config{DeviceType: deviceType}
	if c.Name != "" {
		cfg.PlatformSpecificParams = water.PlatformSpecificParams{
			ComponentID:   "tap0901",
			InterfaceName: c.Name,
			Network:       c.Network,
		}
	}
	return water.New(cfg)
}
