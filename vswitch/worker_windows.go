package vswitch

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/songgao/water"
)

type Bridger struct {

}

type Worker struct {
	WorkerBase
	br          *Bridger
}

func NewWorker(server *TcpServer, c *Config) *Worker {
	w := &Worker{
		br:          nil,
	}

	w.WorkerBase = NewWorkerBase(server, c)
	w.Auth = NewPointAuth(w, c)
	w.Request = NewWithRequest(w, c)
	w.Neighbor = NewNeighber(w, c)
	w.Register()
	w.LoadUsers()

	return w
}

func (w *Worker) NewBr() {
	//TODO
	libol.Warn("Worker.NewBr: TODO")
}

func (w *Worker) NewTap() (*water.Interface, error) {
	//TODO
	libol.Warn("Worker.NewTap: TODO")
	return nil, nil
}

func (w *Worker) Start() {
	w.NewBr()
	w.WorkerBase.Start()
}

func (w *Worker) Stop() {
	w.WorkerBase.Stop()
	if w.br != nil && w.brIp != nil {
		//TODO
	}
}