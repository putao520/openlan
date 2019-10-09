package vswitch

type VSwitch struct {
	worker *Worker
	http   *Http
	State  int
}

func NewVSwitch(c *Config) *VSwitch {
	server := NewTcpServer(c)
	vs := &VSwitch{
		worker: NewWorker(server, c),
		http: nil,
	}
	vs.http = NewHttp(vs.worker, c)
	vs.State = VsInit
	return vs
}

func (vs *VSwitch) Start() {
	if vs.State == VsStarted {
		return
	}
	vs.State = VsStarted

	vs.worker.Start()
	go vs.http.GoStart()
}

func (vs *VSwitch) Stop() {
	if vs.State != VsStarted {
		return
	}

	vs.worker.Stop()
	vs.http.Shutdown()

	vs.State = VsStopped
}

func (vs *VSwitch) GetBrName() string {
	return vs.worker.BrName()
}

func (vs *VSwitch) GetUpTime() int64 {
	return vs.worker.UpTime()
}