package vswitch

import (
	"crypto/tls"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/network"
	"sync"
)

const (
	SW_INIT    = 0x01
	SW_STARTED = 0x02
	SW_STOPPED = 0x03
)

type VSwitch struct {
	Conf *config.VSwitch

	http   *Http
	bridge network.Bridger
	worker *Worker
	lock   sync.RWMutex
	status int
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
		http:   nil,
		status: SW_INIT,
	}

	if c.HttpListen != "" {
		v.http = NewHttp(v.worker, c)
	}

	if c.Bridger == "linux" {
		v.bridge = network.NewLinuxBridge(c.BrName, c.IfMtu)
	} else {
		v.bridge = network.NewVirtualBridge(c.BrName, c.IfMtu)
	}
	if v.bridge.Name() == "" {
		v.bridge.SetName(v.worker.BrName())
	}

	v.worker.Listener = WorkerListener{
		NewTap:  v.NewTap,
		FreeTap: v.FreeTap,
	}

	return &v
}

func (v *VSwitch) Start() bool {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.status == SW_STARTED {
		return false
	} else {
		v.status = SW_STARTED
	}

	v.worker.Start()
	v.bridge.Open(v.Conf.IfAddr)
	if v.http != nil {
		go v.http.Start()
	}
	return true
}

func (v *VSwitch) Stop() bool {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.status != SW_STARTED {
		return false
	} else {
		v.status = SW_STOPPED
	}

	v.bridge.Close()
	v.worker.Stop()
	if v.http != nil {
		v.http.Shutdown()
	}
	return true
}

func (v *VSwitch) IsStated() bool {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.status == SW_STARTED
}

func (v *VSwitch) GetState() string {
	v.lock.Lock()
	defer v.lock.Unlock()

	switch v.status {
	case SW_INIT:
		return "initialized"
	case SW_STARTED:
		return "started"
	case SW_STOPPED:
		return "stopped"
	}

	return ""
}

func (v *VSwitch) GetBrName() string {
	return v.worker.BrName()
}

func (v *VSwitch) GetUpTime() int64 {
	return v.worker.UpTime()
}

func (v *VSwitch) GetWorker() *Worker {
	return v.worker
}

func (v *VSwitch) GetServer() *libol.TcpServer {
	return v.worker.Server
}

func (v *VSwitch) NewTap() (network.Taper, error) {
	var err error
	var dev network.Taper

	libol.Debug("Worker.NewTap")
	if v.Conf.Bridger == "linux" {
		dev, err = network.NewLinuxTap(true, "")
	} else {
		dev, err = network.NewVirtualTap(true, "")
	}
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
