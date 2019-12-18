package vswitch

import (
	"bufio"
	"fmt"
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/models"
	"github.com/lightstar-dev/openlan-go/vswitch/api"
	"github.com/lightstar-dev/openlan-go/vswitch/app"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/point"
)

type Control struct {
	Alias       string
	Server      *libol.TcpServer
	Redis       *libol.RedisCli
	Auth        *app.PointAuth
	Request     *app.WithRequest
	Neighbor    *app.Neighbors
	OnLines     *app.Online
	EnableRedis bool
	Conf        *config.VSwitch

	hooks     []func(*libol.TcpClient, *libol.Frame) error
	usersLock sync.RWMutex
	users     map[string]*models.User
	newTime   int64
	startTime int64
	linksLock sync.RWMutex
	links     map[string]*point.Point
	brName    string
}

func NewControl(server *libol.TcpServer, c *config.VSwitch) *Control {
	w := Control{
		Alias:     c.Alias,
		Server:    server,
		Neighbor:  nil,
		Redis:     nil,
		Conf:      c,
		hooks:     make([]func(*libol.TcpClient, *libol.Frame) error, 0, 64),
		users:     make(map[string]*models.User, 1024),
		newTime:   time.Now().Unix(),
		startTime: 0,
		brName:    c.BrName,
		links:     make(map[string]*point.Point),
	}

	if c.Redis.Enable {
		w.Redis = libol.NewRedisCli(c.Redis.Addr, c.Redis.Auth, c.Redis.Db)
	}

	return &w
}

func (w *Control) GetId() string {
	return w.Server.Addr
}

func (w *Control) String() string {
	return w.GetId()
}

func (w *Control) Init(a api.Worker) {
	w.Auth = app.NewPointAuth(a, w.Conf)
	w.Request = app.NewWithRequest(a, w.Conf)
	w.Neighbor = app.NewNeighbors(a, w.Conf)
	w.OnLines = app.NewOnline(a, w.Conf)

	w.setHook(w.Auth.OnFrame)
	w.setHook(w.Neighbor.OnFrame)
	w.setHook(w.Request.OnFrame)
	w.setHook(w.OnLines.OnFrame)
	w.showHook()

	w.LoadUsers()
}

func (w *Control) LoadUsers() error {
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
			user := models.NewUser(values[0], strings.TrimSpace(values[1]))
			w.AddUser(user)
		}
	}

	return nil
}

func (w *Control) LoadLinks() {
	if w.Conf.Links != nil {
		for _, lc := range w.Conf.Links {
			lc.Default()
			libol.Info("Control.LoadLinks %s", lc)
			w.AddLink(lc)
		}
	}
}

func (w *Control) BrName() string {
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

func (w *Control) showHook() {
	for i, h := range w.hooks {
		libol.Info("Control.showHook k:%d func: %p, %s", i, h, libol.FunName(h))
	}
}

func (w *Control) setHook(hook func(*libol.TcpClient, *libol.Frame) error) {
	w.hooks = append(w.hooks, hook)
}

func (w *Control) onHook(client *libol.TcpClient, data []byte) error {
	frame := libol.NewFrame(data)

	for _, h := range w.hooks {
		libol.Debug("Control.onHook h:%p", h)
		if h != nil {
			if err := h(client, frame); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Control) OnClient(client *libol.TcpClient) error {
	client.SetStatus(libol.CLCONNECTED)

	libol.Info("Control.onClient: %s", client.Addr)

	return nil
}

func (w *Control) OnRecv(client *libol.TcpClient, data []byte) error {
	libol.Debug("Control.onRecv: %s % x", client.Addr, data)

	if err := w.onHook(client, data); err != nil {
		libol.Debug("Control.onRecv: %s dropping by %s", client.Addr, err)
		if client.GetStatus() != libol.CLAUEHED {
			w.Server.DrpCount++
		}
		return err
	}

	point := w.Auth.GetPoint(client)
	if point == nil {
		return libol.Errer("Point not found.")
	}

	dev := point.Device
	if point == nil || point.Device == nil {
		return libol.Errer("Tap devices is nil")
	}

	if _, err := dev.Write(data); err != nil {
		libol.Error("Control.onRecv: %s", err)
		return err
	}

	return nil
}

func (w *Control) OnClose(client *libol.TcpClient) error {
	libol.Info("Control.onClose: %s", client.Addr)

	w.Auth.DelPoint(client)

	return nil
}

func (w *Control) Start() {
	w.startTime = time.Now().Unix()
	if w.Redis != nil {
		if err := w.Redis.Open(); err != nil {
			libol.Error("Control.Start: redis.Open %s", err)
		}
	}

	w.LoadLinks()

	go w.Server.GoAccept()
	go w.Server.GoLoop(w)
}

func (w *Control) Stop() {
	libol.Info("Control.Close")

	w.Server.Close()
	for _, p := range w.links {
		p.Stop()
	}

	w.startTime = 0
}

func (w *Control) AddUser(user *models.User) {
	w.usersLock.Lock()
	defer w.usersLock.Unlock()

	name := user.Name
	if name == "" {
		name = user.Token
	}
	w.users[name] = user
}

func (w *Control) DelUser(name string) {
	w.usersLock.Lock()
	defer w.usersLock.Unlock()

	if _, ok := w.users[name]; ok {
		delete(w.users, name)
	}
}

func (w *Control) GetUser(name string) *models.User {
	w.usersLock.RLock()
	defer w.usersLock.RUnlock()

	if u, ok := w.users[name]; ok {
		return u
	}

	return nil
}

func (w *Control) ListUser() <-chan *models.User {
	c := make(chan *models.User, 128)

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

func (w *Control) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}

func (w *Control) AddLink(c *config.Point) {
	c.Alias = w.Alias
	c.BrName = w.BrName() //Reset bridge name.

	go func() {
		p := point.NewPoint(c)

		w.linksLock.Lock()
		w.links[c.Addr] = p
		w.linksLock.Unlock()

		w.PubLink(p, true)

		p.Start()
	}()
}

func (w *Control) DelLink(addr string) {
	w.linksLock.Lock()
	defer w.linksLock.Unlock()
	if p, ok := w.links[addr]; ok {
		p.Stop()
		delete(w.links, addr)
	}
}

func (w *Control) GetLink(addr string) *point.Point {
	w.linksLock.RLock()
	defer w.linksLock.RUnlock()

	if p, ok := w.links[addr]; ok {
		return p
	}

	return nil
}

func (w *Control) ListLink() <-chan *point.Point {
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

func (w *Control) GetRedis() *libol.RedisCli {
	return w.Redis
}
func (w *Control) GetServer() *libol.TcpServer {
	return w.Server
}

func (w *Control) NewTap() (*models.TapDevice, error) {
	//TODO
	return nil, nil
}

func (w *Control) Send(dev *models.TapDevice, frame []byte) {
	w.Server.TxCount++
}

func (w *Control) PubLink(link *point.Point, isAdd bool) {
	lid := strings.Replace(link.Addr(), ":", "/", -1)
	wid := strings.Replace(w.GetId(), ":", "/", -1)

	key := fmt.Sprintf("%s:link:%s", wid, lid)
	value := map[string]interface{}{
		"remote": link.Addr(),
		"upTime": link.UpTime(),
		"device": link.IfName(),
		"state":  link.State(),
		"isAddr": isAdd,
	}

	if r := w.GetRedis(); r != nil {
		if err := r.HMSet(key, value); err != nil {
			libol.Error("Control.PubLink HMSet %s", err)
		}
	}
}
