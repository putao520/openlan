// +build !linux

package network

func NewBridger(bridge, name string, ifMtu int) Bridger {
	// others platform not support linux bridge.
	return NewVirtualBridge(name, ifMtu)
}
