package network

type Bridger interface {
	Type() string
	Name() string
	SetName(value string)
	Open(addr string)
	Close() error
	AddSlave(dev Taper) error
	DelSlave(dev Taper) error
	Input(m *Framer) error
	SetTimeout(value int)
	Mtu() int
}
