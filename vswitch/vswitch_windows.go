package vswitch

type VSwitch struct {
	worker *Worker
	http   *Http
}

func NewVSwitch(c *Config) *VSwitch {
	//TODO
	return &VSwitch{}
}

func (vs *VSwitch) Start() {
	//TODO
}

func (vs *VSwitch) Stop() {
	//TODO
}

func (vs *VSwitch) GetBrName() string {
	return ""
}

func (vs *VSwitch) GetUpTime() int64 {
	return 0
}
