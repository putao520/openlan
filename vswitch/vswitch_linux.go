package vswitch

import "sync"

type VSwitch struct {
	worker *Worker
	http   *Http

	status  int
	lock    sync.RWMutex
}

func NewVSwitch(c *Config) *VSwitch {
	server := NewTcpServer(c)
	vs := &VSwitch{
		worker: NewWorker(server, c),
		http: nil,
	}
	vs.http = NewHttp(vs.worker, c)
	vs.status = VsInit
	return vs
}

func (vs *VSwitch) Start() {
	vs.lock.Lock()
	defer vs.lock.Unlock()

	if vs.status == VsStarted {
		return
	}
	vs.status = VsStarted

	vs.worker.Start()
	go vs.http.GoStart()
}

func (vs *VSwitch) Stop() {
	vs.lock.Lock()
	defer vs.lock.Unlock()
	
	if vs.status != VsStarted {
		return
	}

	vs.worker.Stop()
	vs.http.Shutdown()

	vs.status = VsStopped
}

func (vs *VSwitch) GetBrName() string {
	return vs.worker.BrName()
}

func (vs *VSwitch) GetUpTime() int64 {
	return vs.worker.UpTime()
}