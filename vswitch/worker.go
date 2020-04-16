package vswitch

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/point"
	"github.com/danieldin95/openlan-go/vswitch/service"
	"sync"
	"time"
)

type Worker struct {
	Alias string
	Conf  config.Bridge

	newTime     int64
	startTime   int64
	linksLock   sync.RWMutex
	links       map[string]*point.Point
	uuid        string
	initialized bool
}

func NewWorker(c config.Bridge) *Worker {
	w := Worker{
		Alias:       c.Alias,
		Conf:        c,
		newTime:     time.Now().Unix(),
		startTime:   0,
		links:       make(map[string]*point.Point),
		initialized: false,
	}

	return &w
}

func (w *Worker) Initialize() {
	w.initialized = true

	_ = service.User.Load(w.Conf.Name, w.Conf.Password)
	_ = service.Network.Load(w.Conf.Name, w.Conf.Network)
}

func (w *Worker) ID() string {
	return w.uuid
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

func (w *Worker) Start(v VSwitcher) {
	libol.Info("Worker.Start: %s", w.Conf.Name)
	if !w.initialized {
		w.Initialize()
	}
	w.uuid = v.UUID()
	w.startTime = time.Now().Unix()
	w.LoadLinks()
}

func (w *Worker) Stop() {
	libol.Info("Worker.Close: %s", w.Conf.Name)
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
	c.BrName = w.Conf.BrName //Reset bridge name.
	c.Allowed = false
	c.Network = w.Conf.Name

	go func() {
		p := point.NewPoint(c)
		p.Initialize()

		w.linksLock.Lock()
		w.links[c.Addr] = p
		w.linksLock.Unlock()

		service.Link.Add(p)
		p.Start()
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
