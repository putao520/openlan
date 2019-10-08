package vswitch

type VSwitch struct {
	worker *Worker
	http   *Http
}

func NewVSwitch(c *Config) *VSwitch {
	//server := NewTcpServer(c)
	//vs := &VSwitch{
	//	worker: NewWorker(server, c),
	//	http: nil,
	//}
	//vs.http = NewHttp(vs.worker, c)
	return &VSwitch{}
}

func (vs *VSwitch) Start() {
	//TODO
}

func (vs *VSwitch) Stop() {
	//TODO
}