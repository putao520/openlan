package vswitch

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/songgao/water"
)

type Worker struct {
	*WorkerBase
	Br *Bridger
}

func NewWorker(server *libol.TcpServer, c *Config) *Worker {
	w := &Worker{
		WorkerBase: NewWorkerBase(server, c),
		Br:         NewBridger(c.BrName, c.IfMtu),
	}
	if w.Br.Name == "" {
		w.Br.Name = w.BrName()
	}

	w.Init(w)
	return w
}

func (w *Worker) NewBr() {
	w.Br.Open(w.Conf.IfAddr)
}

func (w *Worker) FreeBr() {
	w.Br.Close()
}

func (w *Worker) NewTap() (*water.Interface, error) {
	libol.Debug("Worker.newTap")
	dev, err := water.New(water.Config{ DeviceType: water.TAP })
	if err != nil {
		libol.Error("Worker.newTap: %s", err)
		return nil, err
	}

	w.Br.AddSlave(dev.Name())

	libol.Info("Worker.newTap %s", dev.Name())

	return dev, nil
}

func (w *Worker) Start() {
	w.NewBr()
	w.WorkerBase.Start()
}

func (w *Worker) Stop() {
	w.WorkerBase.Stop()
	w.FreeBr()
}
