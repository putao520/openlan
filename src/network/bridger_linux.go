package network

func NewBridger(provider, name string, ifMtu int) Bridger {
	if provider == "virtual" {
		return NewVirtualBridge(name, ifMtu)
	}
	return NewLinuxBridge(name, ifMtu)
}
