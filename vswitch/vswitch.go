package vswitch

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"sync"
)

type Base struct {
	worker *Worker
	http   *Http

	status int
	lock   sync.RWMutex
}

func NewBase(c *Config) Base {
	server := libol.NewTcpServer(c.TcpListen)
	b := Base{
		worker: NewWorker(server, c),
		http:   nil,
	}
	if c.HttpListen != "" {
		b.http = NewHttp(b.worker, c)
	}
	b.status = SWINIT

	return b
}

func (b *Base) Start() bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.status == SWSTARTED {
		return false
	}
	b.status = SWSTARTED

	b.worker.Start()
	if b.http != nil {
		go b.http.GoStart()
	}

	return true
}

func (b *Base) Stop() bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.status != SWSTARTED {
		return false
	}
	b.status = SWSTOPPED

	b.worker.Stop()
	if b.http != nil {
		b.http.Shutdown()
	}

	return true
}

func (b *Base) IsStated() bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.status == SWSTARTED
}

func (b *Base) GetState() string {
	b.lock.Lock()
	defer b.lock.Unlock()

	switch b.status {
	case SWINIT:
		return "initialized"
	case SWSTARTED:
		return "started"
	case SWSTOPPED:
		return "stopped"
	}

	return ""
}

func (b *Base) GetBrName() string {
	return b.worker.BrName()
}

func (b *Base) GetUpTime() int64 {
	return b.worker.UpTime()
}

func (b *Base) GetWorker() *Worker {
	return b.worker
}

func (b *Base) GetServer() *libol.TcpServer {
	return b.worker.Server
}
