package vswitch

import (
	"fmt"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/network"
	"github.com/danieldin95/openlan-go/point"
	"github.com/danieldin95/openlan-go/service"
	"github.com/danieldin95/openlan-go/vswitch/app"
	"strings"
	"sync"
	"time"
)

type WorkerListener struct {
	NewTap  func() (network.Taper, error)
	FreeTap func(dev network.Taper) error
}

type WorkerApps struct {
	Auth     *app.PointAuth
	Request  *app.WithRequest
	Neighbor *app.Neighbors
	OnLines  *app.Online
}

type HookApi func(client *libol.TcpClient, frame *libol.FrameMessage) error

type Worker struct {
	Alias    string
	Conf     *config.VSwitch
	Listener WorkerListener
	Apps     WorkerApps

	server    *libol.TcpServer
	hooks     []HookApi
	newTime   int64
	startTime int64
	linksLock sync.RWMutex
	links     map[string]*point.Point
	brName    string
	uuid      string
}

func NewWorker(server *libol.TcpServer, c *config.VSwitch) *Worker {
	w := Worker{
		Alias:     c.Alias,
		server:    server,
		Conf:      c,
		newTime:   time.Now().Unix(),
		startTime: 0,
		brName:    c.BrName,
		links:     make(map[string]*point.Point),
	}

	return &w
}

func (w *Worker) Initialize() {
	service.User.Load(w.Conf.Password)
	service.Network.Load(w.Conf.Network)

	w.Apps.Auth = app.NewPointAuth(w, w.Conf)
	w.Apps.Request = app.NewWithRequest(w, w.Conf)
	w.Apps.Neighbor = app.NewNeighbors(w, w.Conf)
	w.Apps.OnLines = app.NewOnline(w, w.Conf)

	w.hooks = make([]HookApi, 0, 64)
	w.hooks = append(w.hooks, w.Apps.Auth.OnFrame)
	w.hooks = append(w.hooks, w.Apps.Neighbor.OnFrame)
	w.hooks = append(w.hooks, w.Apps.Request.OnFrame)
	w.hooks = append(w.hooks, w.Apps.OnLines.OnFrame)
	w.ShowHook()
}

func (w *Worker) ID() string {
	return w.server.Addr
}

func (w *Worker) String() string {
	return w.ID()
}

func (w *Worker) LoadLinks() {
	if w.Conf.Links != nil {
		for _, lc := range w.Conf.Links {
			lc.Default()
			w.AddLink(lc)
		}
	}
}

func (w *Worker) BrName() string {
	if w.brName == "" {
		adds := strings.Split(w.server.Addr, ":")
		if len(adds) != 2 {
			w.brName = "brol-default"
		} else {
			w.brName = fmt.Sprintf("brol-%s", adds[1])
		}
	}

	return w.brName
}

func (w *Worker) ShowHook() {
	for i, h := range w.hooks {
		libol.Debug("Worker.showHook k:%d,func:%p,%s", i, h, libol.FunName(h))
	}
}

func (w *Worker) OnHook(client *libol.TcpClient, data []byte) error {
	frame := libol.NewFrameMessage(data)
	for _, h := range w.hooks {
		libol.Debug("Worker.onHook h:%p", h)
		if h != nil {
			if err := h(client, frame); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Worker) OnClient(client *libol.TcpClient) error {
	client.SetStatus(libol.CL_CONNECTED)

	libol.Info("Worker.onClient: %s", client.Addr)

	return nil
}

func (w *Worker) ReadTap(dev network.Taper, readAt func(p []byte) error) {
	defer dev.Close()
	libol.Info("Worker.ReadTap: %s", dev.Name())

	for {
		data := make([]byte, w.Conf.IfMtu)
		n, err := dev.Read(data)
		if err != nil {
			libol.Error("Worker.ReadTap: %s", err)
			break
		}

		libol.Debug("Worker.ReadTap: % x\n", data[:20])
		w.server.Sts.TxCount++
		if err := readAt(data[:n]); err != nil {
			libol.Error("Worker.ReadTap: do-recv %s %s", dev.Name(), err)
		}
	}
}

func (w *Worker) ReadClient(client *libol.TcpClient, data []byte) error {
	libol.Debug("Worker.OnRead: %s % x", client.Addr, data)

	if err := w.OnHook(client, data); err != nil {
		libol.Debug("Worker.OnRead: %s dropping by %s", client.Addr, err)
		if client.Status() != libol.CL_AUEHED {
			w.server.Sts.DrpCount++
		}
		return err
	}

	point := service.Point.Get(client.Addr)
	if point == nil {
		return libol.NewErr("Point not found.")
	}

	dev := point.Device
	if point == nil || point.Device == nil {
		return libol.NewErr("Tap devices is nil")
	}

	if _, err := dev.Write(data); err != nil {
		libol.NewErr("Worker.OnRead: %s", err)
		return err
	}

	return nil
}

func (w *Worker) OnClose(client *libol.TcpClient) error {
	libol.Info("Worker.OnClose: %s", client.Addr)

	service.Point.Del(client.Addr)
	service.Network.FreeAddr(client)

	return nil
}

func (w *Worker) Start(v VSwitcher) {
	w.Initialize()

	w.uuid = v.UUID()
	w.startTime = time.Now().Unix()
	w.LoadLinks()

	go w.server.Accept()
	call := libol.TcpServerListener{
		OnClient: w.OnClient,
		OnClose:  w.OnClose,
		ReadAt:   w.ReadClient,
	}
	go w.server.Loop(call)
}

func (w *Worker) Stop() {
	libol.Info("Worker.Close")

	w.server.Close()
	for _, p := range w.links {
		p.Stop()
	}

	w.startTime = 0
}

func (w *Worker) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}

func (w *Worker) AddLink(c *config.Point) {
	c.Alias = w.Alias
	c.BrName = w.BrName() //Reset bridge name.
	c.Allowed = false

	go func() {
		p := point.NewPoint(c)
		p.Start()

		w.linksLock.Lock()
		w.links[c.Addr] = p
		w.linksLock.Unlock()

		service.Link.Add(p)
	}()
}

func (w *Worker) DelLink(addr string) {
	w.linksLock.Lock()
	defer w.linksLock.Unlock()

	if p, ok := w.links[addr]; ok {
		p.Stop()
		service.Link.Del(p.Addr())
		delete(w.links, addr)
	}
}

func (w *Worker) Server() *libol.TcpServer {
	return w.server
}

func (w *Worker) NewTap() (network.Taper, error) {
	if w.Listener.NewTap == nil {
		return nil, libol.NewErr("Not Implement")
	}
	return w.Listener.NewTap()
}

func (w *Worker) UUID() string {
	return w.uuid
}
