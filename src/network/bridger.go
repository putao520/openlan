package network

type Bridger interface {
	Type() string
	Name() string
	SetName(value string)
	Open(addr string)
	Close() error
	AddSlave(name string) error
	DelSlave(name string) error
	Input(m *Framer) error
	SetTimeout(value int)
	Mtu() int
	Stp(enable bool) error
	Delay(value int) error
}
