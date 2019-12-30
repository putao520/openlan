package vswitch

type Bridger interface {
	Name() string
	SetName(string)
	Open(addr string)
	Close()
	AddSlave(name string) error
}
