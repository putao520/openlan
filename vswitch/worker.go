package vswitch

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/point"
)

type WorkerBase struct {
	Server      *TcpServer
	Auth        *PointAuth
	Request     *WithRequest
	Neighbor    *Neighber
	Redis       *libol.RedisCli
	EnableRedis bool
	Conf        *Config

	brIp        net.IP
	brNet       *net.IPNet
	keys        []int
	hooks       map[int]func(*libol.TcpClient, *libol.Frame) error
	ifMtu       int
	clientsLock sync.RWMutex
	clients     map[*libol.TcpClient]*Point
	usersLock   sync.RWMutex
	users       map[string]*User
	newTime     int64
	startTime   int64
	brName      string
	linksLock   sync.RWMutex
	links       map[string]*point.Point
}

func NewWorkerBase(server *TcpServer, c *Config) WorkerBase {
	w := WorkerBase{
		Server:      server,
		Neighbor:    nil,
		Redis:       libol.NewRedisCli(c.Redis.Addr, c.Redis.Auth, c.Redis.Db),
		EnableRedis: c.Redis.Enable,
		Conf:        c,
		ifMtu:       c.IfMtu,
		hooks:       make(map[int]func(*libol.TcpClient, *libol.Frame) error),
		keys:        make([]int, 0, 1024),
		clients:     make(map[*libol.TcpClient]*Point, 1024),
		users:       make(map[string]*User, 1024),
		newTime:     time.Now().Unix(),
		startTime:   0,
		brName:      c.BrName,
		links:       make(map[string]*point.Point),
	}

	return w
}

func (w *WorkerBase) Register() {
	w.setHook(0x10, w.Neighbor.OnFrame)
	w.setHook(0x00, w.Auth.OnFrame)
	w.setHook(0x01, w.Request.OnFrame)
	w.showHook()
}

func (w *WorkerBase) LoadUsers() error {
	file, err := os.Open(w.Conf.Password)
	if err != nil {
		return err
	}

	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		values := strings.Split(line, ":")
		if len(values) == 2 {
			user := NewUser(values[0], strings.TrimSpace(values[1]))
			w.AddUser(user)
		}
	}

	return nil
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
	for _, k := range w.keys {
		libol.Debug("WorkerBase.showHool k:%d func: %p", k, w.hooks[k])
	}
}

func (w *WorkerBase) setHook(index int, hook func(*libol.TcpClient, *libol.Frame) error) {
	w.hooks[index] = hook
	w.keys = append(w.keys, index)
	sort.Ints(w.keys)
}

func (w *WorkerBase) onHook(client *libol.TcpClient, data []byte) error {
	frame := libol.NewFrame(data)

	for _, k := range w.keys {
		libol.Debug("WorkerBase.onHook k:%d", k)
		if f, ok := w.hooks[k]; ok {
			if err := f(client, frame); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *WorkerBase) handleReq(client *libol.TcpClient, frame *libol.Frame) error {
	return nil
}

func (w *WorkerBase) onClient(client *libol.TcpClient) error {
	client.SetStatus(libol.CLCONNECTED)

	libol.Info("WorkerBase.onClient: %s", client.Addr)

	return nil
}

func (w *WorkerBase) onRecv(client *libol.TcpClient, data []byte) error {
	libol.Debug("WorkerBase.onRecv: %s % x", client.Addr, data)

	if err := w.onHook(client, data); err != nil {
		libol.Debug("WorkerBase.onRecv: %s dropping by %s", client.Addr, err)
		return err
	}

	point := w.GetPoint(client)
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

func (w *WorkerBase) onClose(client *libol.TcpClient) error {
	libol.Info("WorkerBase.onClose: %s", client.Addr)

	w.DelPoint(client)

	return nil
}

func (w *WorkerBase) Start() {
	w.startTime = time.Now().Unix()
	if err := w.Redis.Open(); err != nil {
		libol.Error("WorkerBase.Start: redis.Open %s", err)
	}

	w.LoadLinks()

	go w.Server.GoAccept()
	go w.Server.GoLoop(w.onClient, w.onRecv, w.onClose)
}

func (w *WorkerBase) Stop() {
	libol.Info("WorkerBase.Close")

	w.Server.Close()
	for _, p := range w.links {
		p.Stop()
	}

	w.startTime = 0
}

func (w *WorkerBase) AddUser(user *User) {
	w.usersLock.Lock()
	defer w.usersLock.Unlock()

	name := user.Name
	if name == "" {
		name = user.Token
	}
	w.users[name] = user
}

func (w *WorkerBase) DelUser(name string) {
	w.usersLock.Lock()
	defer w.usersLock.Unlock()

	if _, ok := w.users[name]; ok {
		delete(w.users, name)
	}
}

func (w *WorkerBase) GetUser(name string) *User {
	w.usersLock.RLock()
	defer w.usersLock.RUnlock()

	if u, ok := w.users[name]; ok {
		return u
	}

	return nil
}

func (w *WorkerBase) ListUser() <-chan *User {
	c := make(chan *User, 128)

	go func() {
		w.usersLock.RLock()
		defer w.usersLock.RUnlock()

		for _, u := range w.users {
			c <- u
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (w *WorkerBase) AddPoint(p *Point) {
	w.clientsLock.Lock()
	defer w.clientsLock.Unlock()

	w.PubPoint(p, true)
	w.clients[p.Client] = p
}

func (w *WorkerBase) GetPoint(c *libol.TcpClient) *Point {
	w.clientsLock.RLock()
	defer w.clientsLock.RUnlock()

	if p, ok := w.clients[c]; ok {
		return p
	}
	return nil
}

func (w *WorkerBase) DelPoint(c *libol.TcpClient) {
	w.clientsLock.Lock()
	defer w.clientsLock.Unlock()

	if p, ok := w.clients[c]; ok {
		p.Device.Close()
		w.PubPoint(p, false)
		delete(w.clients, c)
	}
}

func (w *WorkerBase) ListPoint() <-chan *Point {
	c := make(chan *Point, 128)

	go func() {
		w.clientsLock.RLock()
		defer w.clientsLock.RUnlock()

		for _, p := range w.clients {
			c <- p
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (w *WorkerBase) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}

func (w *WorkerBase) PubPoint(p *Point, isAdd bool) {
	if !w.EnableRedis {
		return
	}

	key := fmt.Sprintf("point:%s", strings.Replace(p.Client.String(), ":", "-", -1))
	value := map[string]interface{}{
		"remote":  p.Client.String(),
		"newTime": p.Client.NewTime,
		"device":  p.Device.Name(),
		"active":  isAdd,
	}

	if err := w.Redis.HMSet(key, value); err != nil {
		libol.Error("WorkerBase.PubPoint hset %s", err)
	}
}

func (w *WorkerBase) AddLink(c *point.Config) {
	c.BrName = w.BrName() //Reset bridge name.

	go func() {
		p := point.NewPoint(c)

		w.linksLock.Lock()
		w.links[c.Addr] = p
		w.linksLock.Unlock()

		p.UpLink()
		p.Start()
	}()
}

func (w *WorkerBase) DelLink(addr string) {
	w.linksLock.Lock()
	defer w.linksLock.Unlock()
	if p, ok := w.links[addr]; ok {
		p.Stop()
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
