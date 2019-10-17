package vswitch

import (
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/milosgajdos83/tenus"
	"github.com/songgao/water"
	"net"
)

type Worker struct {
	WorkerBase
	br tenus.Bridger
}

func NewWorker(server *libol.TcpServer, c *Config) *Worker {
	w := &Worker{
		br: nil,
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
	var err error
	var br tenus.Bridger

	addr := w.Conf.IfAddr
	brName := w.BrName()
	br, err = tenus.BridgeFromName(brName)
	if err != nil {
		br, err = tenus.NewBridgeWithName(brName)
		if err != nil {
			libol.Error("Worker.newBr: %s", err)
		}
	}

	brCtl := libol.NewBrCtl(brName)
	if err := brCtl.Stp(true); err != nil {
		libol.Error("Worker.newBr.Stp: %s", err)
	}

	if err = br.SetLinkUp(); err != nil {
		libol.Error("Worker.newBr: %s", err)
	}

	libol.Info("Worker.newBr %s", brName)

	if addr != "" {
		ip, net, err := net.ParseCIDR(addr)
		if err != nil {
			libol.Error("Worker.newBr.ParseCIDR %s : %s", addr, err)
		}
		if err := br.SetLinkIp(ip, net); err != nil {
			libol.Error("Worker.newBr.SetLinkIp %s : %s", brName, err)
		}

		w.brIp = ip
		w.brNet = net
	}

	w.br = br
}

func (w *Worker) NewTap() (*water.Interface, error) {
	libol.Debug("Worker.newTap")
	dev, err := water.New(water.Config{
		DeviceType: water.TAP,
	})
	if err != nil {
		libol.Error("Worker.newTap: %s", err)
		return nil, err
	}

	link, err := tenus.NewLinkFrom(dev.Name())
	if err != nil {
		libol.Error("Worker.newTap: Get dev %s: %s", dev.Name(), err)
		return nil, err
	}

	if err := link.SetLinkUp(); err != nil {
		libol.Error("Worker.newTap: ", err)
	}

	if err := w.br.AddSlaveIfc(link.NetInterface()); err != nil {
		libol.Error("Worker.newTap: Switch dev %s: %s", dev.Name(), err)
		return nil, err
	}

	libol.Info("Worker.newTap %s", dev.Name())

	return dev, nil
}

func (w *Worker) Start() {
	w.NewBr()
	w.WorkerBase.Start()
}

func (w *Worker) Stop() {
	w.WorkerBase.Stop()
	if w.br != nil && w.brIp != nil {
		if err := w.br.UnsetLinkIp(w.brIp, w.brNet); err != nil {
			libol.Error("Worker.Close.UnsetLinkIp %s : %s", w.br.NetInterface().Name, err)
		}
	}
}
