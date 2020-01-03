package network

type Bridger interface {
	Name() string
	SetName(value string)
	Open(addr string)
	Close() error
	AddSlave(dev Taper) error
	DelSlave(dev Taper) error
	Input(m *Framer) error
}
