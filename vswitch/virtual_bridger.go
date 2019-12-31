package vswitch

type VirtualBridger struct {
	mtu    int
	name   string
}

func NewVirtualBridger(name string, mtu int) *VirtualBridger {
	b := &VirtualBridger{
		name: name,
		mtu:  mtu,
	}
	return b
}

func (b *VirtualBridger) Open(addr string) {
}

func (b *VirtualBridger) Close() {
}

func (b *VirtualBridger) AddSlave(name string) error {
	return nil
}


func (b *VirtualBridger) Name() string {
	return b.name
}

func (b *VirtualBridger) SetName(value string)  {
	b.name = value
}
