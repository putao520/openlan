package vswitch

import "sync"

type VSwitch struct {
	worker *Worker
	http   *Http

	status int
	lock   sync.RWMutex
}

func NewVSwitch(c *Config) *VSwitch {
	server := NewTcpServer(c)
	vs := &VSwitch{
		worker: NewWorker(server, c),
		http:   nil,
	}
	if c.HttpListen != "" {
		vs.http = NewHttp(vs.worker, c)
	}
	vs.status = VSINIT
	return vs
}

func (vs *VSwitch) Start() {
	vs.lock.Lock()
	defer vs.lock.Unlock()

	if vs.status == VSSTARTED {
		return
	}
	vs.status = VSSTARTED

	vs.worker.Start()
	if vs.http != nil {
		go vs.http.GoStart()
	}
}

func (vs *VSwitch) Stop() {
	vs.lock.Lock()
	defer vs.lock.Unlock()

	if vs.status != VSSTARTED {
		return
	}
	vs.status = VSTOPPED

	vs.worker.Stop()
	if vs.http != nil {
		vs.http.Shutdown()
	}
}

func (vs *VSwitch) GetBrName() string {
	return vs.worker.BrName()
}

func (vs *VSwitch) GetUpTime() int64 {
	return vs.worker.UpTime()
}
