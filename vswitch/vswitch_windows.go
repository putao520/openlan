package vswitch

type VSwitch struct {
	worker *Worker
	http   *Http
}

func NewVSwitch(c *Config) *VSwitch {
	vs := &VSwitch{}
	vs.Base = NewBase(c)

	return vs
}

func (vs *VSwitch) Start() {
	//TODO
}

func (vs *VSwitch) Stop() {
	//TODO
}
