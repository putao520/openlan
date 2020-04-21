package ctl

type Listener interface {
	GetCtl(id, data string) error
	AddCtl(id, data string) error
	DelCtl(id, data string) error
	ModCtl(id, data string) error
}
