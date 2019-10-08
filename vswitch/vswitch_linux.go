package vswitch

type VSwitch struct {
	worker *Worker
	http   *Http
}

func NewVSwitch(c *Config) (*VSwitch) {
	server := NewTcpServer(c)
	vs := &VSwitch{
		worker: NewWorker(server, c),
		http: nil,
	}
	vs.http = NewHttp(vs.worker, c)
	return vs
}

func (vs *VSwitch) Start() {
	vs.worker.Start()
	go vs.http.GoStart()
}

func (vs *VSwitch) Stop() {
	vs.worker.Stop()
}