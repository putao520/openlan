package network

type VirBridge struct {
	mtu  int
	name string
}

func NewVirBridge(name string, mtu int) *VirBridge {
	b := &VirBridge{
		name: name,
		mtu:  mtu,
	}
	return b
}

func (b *VirBridge) Open(addr string) {
}

func (b *VirBridge) Close() {
}

func (b *VirBridge) AddSlave(name string) error {
	return nil
}

func (b *VirBridge) Name() string {
	return b.name
}

func (b *VirBridge) SetName(value string) {
	b.name = value
}
