package vswitch

import (
	"crypto/tls"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/network"
	"sync"
)

type VSwitcher interface {
	UUID() string
}

type VSwitch struct {
	Conf *config.VSwitch

	http   *Http
	bridge network.Bridger
	worker *Worker
	lock   sync.RWMutex
	uuid   string
}

func NewVSwitch(c *config.VSwitch) *VSwitch {
	var tlsConf *tls.Config

	if c.KeyFile != "" && c.CrtFile != "" {
		cer, err := tls.LoadX509KeyPair(c.CrtFile, c.KeyFile)
		if err != nil {
			libol.Error("NewVSwitch: %s", err)
		}
		tlsConf = &tls.Config{Certificates: []tls.Certificate{cer}}
	}

	server := libol.NewTcpServer(c.TcpListen, tlsConf)
	v := VSwitch{
		Conf:   c,
		worker: NewWorker(server, c),
	}

	return &v
}

func (v *VSwitch) Initialize() {
	if v.Conf.HttpListen != "" {
		v.http = NewHttp(v.worker, v.Conf)
	}

	v.bridge = network.NewBridger(v.Conf.Bridger, v.Conf.BrName, v.Conf.IfMtu)
	if v.bridge.Name() == "" {
		v.bridge.SetName(v.worker.BrName())
	}

	v.worker.Listener = WorkerListener{
		NewTap:  v.NewTap,
		FreeTap: v.FreeTap,
	}
}

func (v *VSwitch) Start() error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.bridge != nil || v.http != nil {
		return libol.NewErr("already running")
	}

	v.Initialize()
	v.worker.Start(v)
	v.bridge.Open(v.Conf.IfAddr)
	if v.http != nil {
		go v.http.Start()
	}
	return nil
}

func (v *VSwitch) Stop() error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.bridge == nil {
		return libol.NewErr("already closed")
	}

	v.bridge.Close()
	v.bridge = nil
	if v.http != nil {
		v.http.Shutdown()
		v.http = nil
	}
	v.worker.Stop()
	return nil
}

func (v *VSwitch) BrName() string {
	return v.worker.BrName()
}

func (v *VSwitch) UpTime() int64 {
	return v.worker.UpTime()
}

func (v *VSwitch) Worker() *Worker {
	return v.worker
}

func (v *VSwitch) Server() *libol.TcpServer {
	return v.worker.Server()
}

func (v *VSwitch) NewTap() (network.Taper, error) {
	libol.Debug("Worker.NewTap")

	dev, err := network.NewTaper(v.Conf.Bridger, "", true)
	if err != nil {
		libol.Error("Worker.NewTap: %s", err)
		return nil, err
	}

	dev.Up()
	v.bridge.AddSlave(dev)
	libol.Info("Worker.NewTap %s", dev.Name())

	return dev, nil
}

func (v *VSwitch) FreeTap(dev network.Taper) error {
	v.bridge.DelSlave(dev)
	libol.Info("Worker.FreeTap %s", dev.Name())

	return nil
}

func (v *VSwitch) UUID() string {
	if v.uuid == "" {
		v.uuid = libol.GenToken(32)
	}
	return v.uuid
}
