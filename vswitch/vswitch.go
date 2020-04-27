package vswitch

import (
	"crypto/tls"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/network"
	"github.com/danieldin95/openlan-go/vswitch/app"
	"github.com/danieldin95/openlan-go/vswitch/ctrls"
	"github.com/danieldin95/openlan-go/vswitch/service"
	"sync"
	"time"
)

type Apps struct {
	Auth     *app.PointAuth
	Request  *app.WithRequest
	Neighbor *app.Neighbors
	OnLines  *app.Online
}

type Hook func(client *libol.TcpClient, frame *libol.FrameMessage) error

type VSwitch struct {
	Conf config.VSwitch
	Apps Apps

	hooks      []Hook
	http       *Http
	server     *libol.TcpServer
	bridge     map[string]network.Bridger
	worker     map[string]*Worker
	lock       sync.RWMutex
	uuid       string
	newTime    int64
	initialize bool
}

func NewVSwitch(c config.VSwitch) *VSwitch {
	var tlsConf *tls.Config

	if c.KeyFile != "" && c.CrtFile != "" {
		cer, err := tls.LoadX509KeyPair(c.CrtFile, c.KeyFile)
		if err != nil {
			libol.Error("NewVSwitch: %s", err)
		}
		tlsConf = &tls.Config{Certificates: []tls.Certificate{cer}}
	}

	v := VSwitch{
		Conf:       c,
		worker:     make(map[string]*Worker, 32),
		bridge:     make(map[string]network.Bridger, 32),
		server:     libol.NewTcpServer(c.Listen, tlsConf),
		newTime:    time.Now().Unix(),
		initialize: false,
	}
	return &v
}

func (v *VSwitch) Initialize() {
	v.initialize = true
	if v.Conf.Http != nil {
		v.http = NewHttp(v, v.Conf)
	}
	for _, nCfg := range v.Conf.Network {
		name := nCfg.Name
		brCfg := nCfg.Bridge
		v.worker[name] = NewWorker(*nCfg)
		v.bridge[name] = network.NewBridger(brCfg.Provider, brCfg.Name, brCfg.Mtu)
	}

	v.Apps.Auth = app.NewPointAuth(v, v.Conf)
	v.Apps.Request = app.NewWithRequest(v, v.Conf)
	v.Apps.Neighbor = app.NewNeighbors(v, v.Conf)
	v.Apps.OnLines = app.NewOnline(v, v.Conf)

	v.hooks = make([]Hook, 0, 64)
	v.hooks = append(v.hooks, v.Apps.Auth.OnFrame)
	v.hooks = append(v.hooks, v.Apps.Neighbor.OnFrame)
	v.hooks = append(v.hooks, v.Apps.Request.OnFrame)
	v.hooks = append(v.hooks, v.Apps.OnLines.OnFrame)
	for i, h := range v.hooks {
		libol.Debug("Worker.showHook: k %d, func %p, %s", i, h, libol.FunName(h))
	}

	// Controller
	ctrls.Load(v.Conf.ConfDir + "/ctrl.json")
	if ctrls.Ctrl.Name == "" {
		ctrls.Ctrl.Name = v.Conf.Alias
	}
	ctrls.Ctrl.Switcher = v
}

func (v *VSwitch) OnHook(client *libol.TcpClient, data []byte) error {
	frame := libol.NewFrameMessage(data)
	for _, h := range v.hooks {
		libol.Log("Worker.onHook: h %p", h)
		if h != nil {
			if err := h(client, frame); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *VSwitch) OnClient(client *libol.TcpClient) error {
	client.SetStatus(libol.CL_CONNECTED)
	libol.Info("VSwitch.onClient: %s", client.Addr)
	return nil
}

func (v *VSwitch) ReadClient(client *libol.TcpClient, data []byte) error {
	libol.Log("VSwitch.ReadClient: %s % x", client.Addr, data)
	if err := v.OnHook(client, data); err != nil {
		libol.Debug("VSwitch.OnRead: %s dropping by %s", client.Addr, err)
		if client.Status() != libol.CL_AUEHED {
			v.server.Sts.DrpCount++
		}
		return nil
	}

	private := client.Private()
	if private != nil {
		point := private.(*models.Point)
		dev := point.Device
		if point == nil || dev == nil {
			return libol.NewErr("Tap devices is nil")
		}
		if _, err := dev.Write(data); err != nil {
			libol.Error("Worker.OnRead: %s", err)
			return err
		}
		return nil
	}
	return libol.NewErr("%s Point not found.", client)
}

func (v *VSwitch) OnClose(client *libol.TcpClient) error {
	libol.Info("VSwitch.OnClose: %s", client.Addr)

	service.Point.Del(client.Addr)
	service.Network.FreeAddr(client)

	return nil
}

func (v *VSwitch) Start() error {
	v.lock.Lock()
	defer v.lock.Unlock()

	libol.Debug("VSwitch.Start")
	if !v.initialize {
		v.Initialize()
	}

	for _, nCfg := range v.Conf.Network {
		if br, ok := v.bridge[nCfg.Name]; ok {
			brCfg := nCfg.Bridge
			br.Open(brCfg.Address)
		}
	}
	go v.server.Accept()
	call := libol.TcpServerListener{
		OnClient: v.OnClient,
		OnClose:  v.OnClose,
		ReadAt:   v.ReadClient,
	}
	go v.server.Loop(call)
	for _, w := range v.worker {
		w.Start(v)
	}
	if v.http != nil {
		go v.http.Start()
	}
	go ctrls.Ctrl.Start()

	return nil
}

func (v *VSwitch) Stop() error {
	v.lock.Lock()
	defer v.lock.Unlock()

	libol.Debug("VSwitch.Stop")
	ctrls.Ctrl.Stop()
	if v.bridge == nil {
		return libol.NewErr("already closed")
	}
	for _, nCfg := range v.Conf.Network {
		if br, ok := v.bridge[nCfg.Name]; ok {
			brCfg := nCfg.Bridge
			_ = br.Close()
			delete(v.bridge, brCfg.Name)
		}
	}
	if v.http != nil {
		v.http.Shutdown()
		v.http = nil
	}
	v.server.Close()
	for _, w := range v.worker {
		w.Stop()
	}
	return nil
}

func (v *VSwitch) Alias() string {
	return v.Conf.Alias
}

func (v *VSwitch) UpTime() int64 {
	return time.Now().Unix() - v.newTime
}

func (v *VSwitch) Server() *libol.TcpServer {
	return v.server
}

func (v *VSwitch) NewTap(tenant string) (network.Taper, error) {
	v.lock.RLock()
	defer v.lock.RUnlock()
	libol.Debug("Worker.NewTap")

	br, ok := v.bridge[tenant]
	if !ok {
		return nil, libol.NewErr("Not found bridge %s", tenant)
	}
	dev, err := network.NewTaper(br.Type(), tenant, network.TapConfig{Type: network.TAP})
	if err != nil {
		libol.Error("Worker.NewTap: %s", err)
		return nil, err
	}
	mtu := br.Mtu()
	dev.SetMtu(mtu)
	dev.Up()
	_ = br.AddSlave(dev)
	libol.Info("Worker.NewTap: %s on %s", dev.Name(), tenant)
	return dev, nil
}

func (v *VSwitch) FreeTap(dev network.Taper) error {
	br, ok := v.bridge[dev.Tenant()]
	if !ok {
		return libol.NewErr("Not found bridge %s", dev.Tenant())
	}
	_ = br.DelSlave(dev)
	libol.Info("Worker.FreeTap: %s", dev.Name())
	return nil
}

func (v *VSwitch) UUID() string {
	if v.uuid == "" {
		v.uuid = libol.GenToken(32)
	}
	return v.uuid
}

func (v *VSwitch) AddLink(tenant string, c *config.Point) {

}

func (v *VSwitch) DelLink(tenant, addr string) {

}

func (v *VSwitch) ReadTap(dev network.Taper, readAt func(p []byte) error) {
	defer dev.Close()
	libol.Info("VSwitch.ReadTap: %s", dev.Name())

	for {
		data := make([]byte, dev.Mtu())
		n, err := dev.Read(data)
		if err != nil {
			libol.Error("VSwitch.ReadTap: %s", err)
			break
		}
		libol.Log("VSwitch.ReadTap: % x\n", data)
		if err := readAt(data[:n]); err != nil {
			libol.Error("VSwitch.ReadTap: do-recv %s %s", dev.Name(), err)
			break
		}
	}
}

func (v *VSwitch) Config() *config.VSwitch {
	return &v.Conf
}
