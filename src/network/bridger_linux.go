package network

func NewBridger(bridge, name string, ifMtu int) Bridger {
	if bridge == "linux" {
		return NewLinuxBridge(name, ifMtu)
	}
	return NewVirtualBridge(name, ifMtu)
}
