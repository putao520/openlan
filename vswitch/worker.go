package vswitch

import (
	"fmt"
	"github.com/danieldin95/openlan-go/service"
	"github.com/songgao/water"
	"strings"
	"sync"
	"time"

	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/point"
	"github.com/danieldin95/openlan-go/vswitch/api"
	"github.com/danieldin95/openlan-go/vswitch/app"
)

type WorkerBase struct {
	Alias       string
	Server      *libol.TcpServer
	Auth        *app.PointAuth
	Request     *app.WithRequest
	Neighbor    *app.Neighbors
	OnLines     *app.Online
	Conf        *config.VSwitch

	hooks     []func(*libol.TcpClient, *libol.Frame) error
	newTime   int64
	startTime int64
	linksLock sync.RWMutex
	links     map[string]*point.Point
	brName    string
}

func NewWorkerBase(server *libol.TcpServer, c *config.VSwitch) *WorkerBase {
	w := WorkerBase{
		Alias:     c.Alias,
		Server:    server,
		Neighbor:  nil,
		Conf:      c,
		hooks:     make([]func(*libol.TcpClient, *libol.Frame) error, 0, 64),
		newTime:   time.Now().Unix(),
		startTime: 0,
		brName:    c.BrName,
		links:     make(map[string]*point.Point),
	}

	service.User.Load(w.Conf.Password)
	service.Network.Load(w.Conf.Network)
	service.Storage.Open(c.Redis.Addr, c.Redis.Auth, c.Redis.Db)

	return &w
}

func (w *WorkerBase) GetId() string {
	return w.Server.Addr
}

func (w *WorkerBase) String() string {
	return w.GetId()
}

func (w *WorkerBase) Init(a api.Worker) {
	w.Auth = app.NewPointAuth(a, w.Conf)
	w.Request = app.NewWithRequest(a, w.Conf)
	w.Neighbor = app.NewNeighbors(a, w.Conf)
	w.OnLines = app.NewOnline(a, w.Conf)

	w.setHook(w.Auth.OnFrame)
	w.setHook(w.Neighbor.OnFrame)
	w.setHook(w.Request.OnFrame)
	w.setHook(w.OnLines.OnFrame)
	w.showHook()
}

func (w *WorkerBase) LoadLinks() {
	if w.Conf.Links != nil {
		for _, lc := range w.Conf.Links {
			lc.Default()
			libol.Info("WorkerBase.LoadLinks %s", lc)
			w.AddLink(lc)
		}
	}
}

func (w *WorkerBase) BrName() string {
	if w.brName == "" {
		adds := strings.Split(w.Server.Addr, ":")
		if len(adds) != 2 {
			w.brName = "brol-default"
		} else {
			w.brName = fmt.Sprintf("brol-%s", adds[1])
		}
	}

	return w.brName
}

func (w *WorkerBase) showHook() {
	for i, h := range w.hooks {
		libol.Info("WorkerBase.showHook k:%d func: %p, %s", i, h, libol.FunName(h))
	}
}

func (w *WorkerBase) setHook(hook func(*libol.TcpClient, *libol.Frame) error) {
	w.hooks = append(w.hooks, hook)
}

func (w *WorkerBase) onHook(client *libol.TcpClient, data []byte) error {
	frame := libol.NewFrame(data)

	for _, h := range w.hooks {
		libol.Debug("WorkerBase.onHook h:%p", h)
		if h != nil {
			if err := h(client, frame); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *WorkerBase) OnClient(client *libol.TcpClient) error {
	client.SetStatus(libol.CLCONNECTED)

	libol.Info("WorkerBase.onClient: %s", client.Addr)

	return nil
}

func (w *WorkerBase) OnRecv(client *libol.TcpClient, data []byte) error {
	libol.Debug("WorkerBase.onRecv: %s % x", client.Addr, data)

	if err := w.onHook(client, data); err != nil {
		libol.Debug("WorkerBase.onRecv: %s dropping by %s", client.Addr, err)
		if client.GetStatus() != libol.CLAUEHED {
			w.Server.DrpCount++
		}
		return err
	}

	point := service.Point.GetPoint(client)
	if point == nil {
		return libol.Errer("Point not found.")
	}

	dev := point.Device
	if point == nil || point.Device == nil {
		return libol.Errer("Tap devices is nil")
	}

	if _, err := dev.Write(data); err != nil {
		libol.Error("WorkerBase.onRecv: %s", err)
		return err
	}

	return nil
}

func (w *WorkerBase) OnClose(client *libol.TcpClient) error {
	libol.Info("WorkerBase.onClose: %s", client.Addr)

	service.Point.DelPoint(client)
	service.Network.FreeAddr(client)

	return nil
}

func (w *WorkerBase) Start() {
	w.startTime = time.Now().Unix()

	w.LoadLinks()

	go w.Server.GoAccept()
	go w.Server.GoLoop(w)
}

func (w *WorkerBase) Stop() {
	libol.Info("WorkerBase.Close")

	w.Server.Close()
	for _, p := range w.links {
		p.Stop()
	}

	w.startTime = 0
}

func (w *WorkerBase) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}

func (w *WorkerBase) AddLink(c *config.Point) {
	c.Alias = w.Alias
	c.BrName = w.BrName() //Reset bridge name.

	go func() {
		p := point.NewPoint(c)

		w.linksLock.Lock()
		w.links[c.Addr] = p
		w.linksLock.Unlock()

		service.Storage.SaveLink(w.GetId(), p, true)
		p.Start()
	}()
}

func (w *WorkerBase) DelLink(addr string) {
	w.linksLock.Lock()
	defer w.linksLock.Unlock()
	if p, ok := w.links[addr]; ok {
		p.Stop()
		service.Storage.SaveLink(w.GetId(), p, false)
		delete(w.links, addr)
	}
}

func (w *WorkerBase) GetLink(addr string) *point.Point {
	w.linksLock.RLock()
	defer w.linksLock.RUnlock()

	if p, ok := w.links[addr]; ok {
		return p
	}

	return nil
}

func (w *WorkerBase) ListLink() <-chan *point.Point {
	c := make(chan *point.Point, 128)

	go func() {
		w.linksLock.RLock()
		defer w.linksLock.RUnlock()

		for _, p := range w.links {
			c <- p
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (w *WorkerBase) GetServer() *libol.TcpServer {
	return w.Server
}

func (w *WorkerBase) NewTap() (*models.TapDevice, error) {
	//TODO
	return nil, nil
}

func (w *WorkerBase) Send(dev *models.TapDevice, frame []byte) {
	w.Server.TxCount++
}

type Worker struct {
	*WorkerBase
	Br *Bridger
}

func NewWorker(server *libol.TcpServer, c *config.VSwitch) *Worker {
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

func (w *Worker) NewTap() (*models.TapDevice, error) {
	libol.Debug("Worker.newTap")
	dev, err := water.New(water.Config{DeviceType: water.TAP})
	if err != nil {
		libol.Error("Worker.newTap: %s", err)
		return nil, err
	}

	w.Br.AddSlave(dev.Name())

	libol.Info("Worker.newTap %s", dev.Name())

	return models.NewTapDevice(dev), nil
}

func (w *Worker) Start() {
	w.NewBr()
	w.WorkerBase.Start()
}

func (w *Worker) Stop() {
	w.WorkerBase.Stop()
	w.FreeBr()
}
