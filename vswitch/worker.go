package vswitch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/point"
	"github.com/milosgajdos83/tenus"
	"github.com/songgao/water"
)

type Point struct {
	Client *libol.TcpClient
	Device *water.Interface
}

func NewPoint(c *libol.TcpClient, d *water.Interface) (w *Point) {
	w = &Point{
		Client: c,
		Device: d,
	}

	return
}

type Worker struct {
	Server      *TcpServer
	Auth        *PointAuth
	Request     *WithRequest
	Neighbor    *Neighber
	Redis       *libol.RedisCli
	EnableRedis bool
	Conf        *Config

	br          tenus.Bridger
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
	brName      string
	linksLock   sync.RWMutex
	links       map[string]*point.Point
}

func NewWorker(server *TcpServer, c *Config) (w *Worker) {
	w = &Worker{
		Server:      server,
		Neighbor:    nil,
		Redis:       libol.NewRedisCli(c.Redis.Addr, c.Redis.Auth, c.Redis.Db),
		EnableRedis: c.Redis.Enable,
		Conf:        c,
		br:          nil,
		ifMtu:       c.IfMtu,
		hooks:       make(map[int]func(*libol.TcpClient, *libol.Frame) error),
		keys:        make([]int, 0, 1024),
		clients:     make(map[*libol.TcpClient]*Point, 1024),
		users:       make(map[string]*User, 1024),
		newTime:     time.Now().Unix(),
		brName:      c.BrName,
		links:       make(map[string]*point.Point),
	}

	w.Auth = NewPointAuth(w, c)
	w.Request = NewWithRequest(w, c)
	w.Neighbor = NewNeighber(w, c)
	w.Register()
	w.LoadUsers()

	return
}

func (w *Worker) Register() {
	w.setHook(0x10, w.Neighbor.OnFrame)
	w.setHook(0x00, w.Auth.OnFrame)
	w.setHook(0x01, w.Request.OnFrame)
	w.showHook()
}

func (w *Worker) LoadUsers() error {
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

func (w *Worker) LoadLinks() {
	if w.Conf.Links != nil {
		for _, lc := range w.Conf.Links {
			lc.Default()
			libol.Info("Worker.LoadLinks %s", lc)
			w.AddLink(lc)
		}
	}
}

func (w *Worker) BrName() string {
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
	if err := w.Redis.Open(); err != nil {
		libol.Error("Worker.Start: redis.Open %s", err)
	}

	w.NewBr()
	w.LoadLinks()

	go w.Server.GoAccept()
	go w.Server.GoLoop(w.onClient, w.onRecv, w.onClose)
}

func (w *Worker) showHook() {
	for _, k := range w.keys {
		libol.Debug("Worker.showHool k:%d func: %p", k, w.hooks[k])
	}
}

func (w *Worker) setHook(index int, hook func(*libol.TcpClient, *libol.Frame) error) {
	w.hooks[index] = hook
	w.keys = append(w.keys, index)
	sort.Ints(w.keys)
}

func (w *Worker) onHook(client *libol.TcpClient, data []byte) error {
	frame := libol.NewFrame(data)

	for _, k := range w.keys {
		libol.Debug("Worker.onHook k:%d", k)
		if f, ok := w.hooks[k]; ok {
			if err := f(client, frame); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Worker) handleReq(client *libol.TcpClient, frame *libol.Frame) error {
	return nil
}

func (w *Worker) onClient(client *libol.TcpClient) error {
	client.SetStatus(libol.ClConnected)

	libol.Info("Worker.onClient: %s", client.Addr)

	return nil
}

func (w *Worker) onRecv(client *libol.TcpClient, data []byte) error {
	//TODO Hook packets such as ARP Learning.
	libol.Debug("Worker.onRecv: %s % x", client.Addr, data)

	if err := w.onHook(client, data); err != nil {
		libol.Debug("Worker.onRecv: %s dropping by %s", client.Addr, err)
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
		libol.Error("Worker.onRecv: %s", err)
		return err
	}

	return nil
}

func (w *Worker) onClose(client *libol.TcpClient) error {
	libol.Info("Worker.onClose: %s", client.Addr)

	w.DelPoint(client)

	return nil
}

func (w *Worker) Stop() {
	libol.Info("Worker.Close")

	w.Server.Close()

	if w.br != nil && w.brIp != nil {
		if err := w.br.UnsetLinkIp(w.brIp, w.brNet); err != nil {
			libol.Error("Worker.Close.UnsetLinkIp %s : %s", w.br.NetInterface().Name, err)
		}
	}

	for _, p := range w.links {
		p.Stop()
	}
}

func (w *Worker) AddUser(user *User) {
	w.usersLock.Lock()
	defer w.usersLock.Unlock()

	name := user.Name
	if name == "" {
		name = user.Token
	}
	w.users[name] = user
}

func (w *Worker) GetUser(name string) *User {
	w.usersLock.RLock()
	defer w.usersLock.RUnlock()

	if u, ok := w.users[name]; ok {
		return u
	}

	return nil
}

func (w *Worker) ListUser() <-chan *User {
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

func (w *Worker) AddPoint(p *Point) {
	w.clientsLock.Lock()
	defer w.clientsLock.Unlock()

	w.PubPoint(p, true)
	w.clients[p.Client] = p
}

func (w *Worker) GetPoint(c *libol.TcpClient) *Point {
	w.clientsLock.RLock()
	defer w.clientsLock.RUnlock()

	if p, ok := w.clients[c]; ok {
		return p
	}
	return nil
}

func (w *Worker) DelPoint(c *libol.TcpClient) {
	w.clientsLock.Lock()
	defer w.clientsLock.Unlock()

	if p, ok := w.clients[c]; ok {
		p.Device.Close()
		w.PubPoint(p, false)
		delete(w.clients, c)
	}
}

func (w *Worker) ListPoint() <-chan *Point {
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

func (w *Worker) UpTime() int64 {
	return time.Now().Unix() - w.newTime
}

func (w *Worker) PubPoint(p *Point, isadd bool) {
	if !w.EnableRedis {
		return
	}

	key := fmt.Sprintf("point:%s", strings.Replace(p.Client.String(), ":", "-", -1))
	value := map[string]interface{}{
		"remote":  p.Client.String(),
		"newTime": p.Client.NewTime,
		"device":  p.Device.Name(),
		"active": isadd,
	}

	if err := w.Redis.HMSet(key, value); err != nil {
		libol.Error("Worker.PubPoint hset %s", err)
	}
}

func (w *Worker) AddLink(c *point.Config) {
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

func (w *Worker) DelLink(addr string) {
	w.linksLock.Lock()
	defer w.linksLock.Unlock()
	if p, ok := w.links[addr]; ok {
		p.Stop()
		delete(w.links, addr)
	}
}

func (w *Worker) GetLink(addr string) *point.Point {
	w.linksLock.RLock()
	defer w.linksLock.RUnlock()

	if p, ok := w.links[addr]; ok {
		return p
	}

	return nil
}

func (w *Worker) ListLink() <-chan *point.Point {
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

type PointAuth struct {
	ifMtu  int
	worker *Worker
}

func NewPointAuth(worker *Worker, c *Config) (w *PointAuth) {
	w = &PointAuth{
		ifMtu:  c.IfMtu,
		worker: worker,
	}
	return
}

func (w *PointAuth) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("PointAuth.OnFrame % x.", frame.Data)

	if libol.IsInst(frame.Data) {
		action := libol.DecAction(frame.Data)
		libol.Debug("PointAuth.OnFrame.action: %s", action)

		if action == "logi=" {
			if err := w.handleLogin(client, libol.DecBody(frame.Data)); err != nil {
				libol.Error("PointAuth.OnFrame: %s", err)
				client.SendResp("login", err.Error())
				client.Close()
				return err
			}
			client.SendResp("login", "okay.")
		}

		return nil
	}

	if client.GetStatus() != libol.ClAuthed {
		client.Dropped++
		libol.Debug("PointAuth.onRecv: %s unauth", client.Addr)
		return libol.Errer("Unauthed client.")
	}

	return nil
}

func (w *PointAuth) handleLogin(client *libol.TcpClient, data string) error {
	libol.Debug("PointAuth.handleLogin: %s", data)

	if client.GetStatus() == libol.ClAuthed {
		libol.Warn("PointAuth.handleLogin: already authed %s", client)
		return nil
	}

	user := NewUser("", "")
	if err := json.Unmarshal([]byte(data), user); err != nil {
		return libol.Errer("Invalid json data.")
	}

	name := user.Name
	if user.Token != "" {
		name = user.Token
	}
	_user := w.worker.GetUser(name)
	if _user != nil {
		if _user.Password == user.Password {
			client.SetStatus(libol.ClAuthed)
			libol.Info("PointAuth.handleLogin: %s Authed", client.Addr)
			w.onAuth(client)
			return nil
		}

		client.SetStatus(libol.ClUnauth)
	}

	return libol.Errer("Auth failed.")
}

func (w *PointAuth) onAuth(client *libol.TcpClient) error {
	if client.GetStatus() != libol.ClAuthed {
		return libol.Errer("not authed.")
	}

	libol.Info("PointAuth.onAuth: %s", client.Addr)
	dev, err := w.worker.NewTap()
	if err != nil {
		return err
	}

	p := NewPoint(client, dev)
	w.worker.AddPoint(p)

	go w.GoRecv(dev, client.SendMsg)

	return nil
}

func (w *PointAuth) GoRecv(dev *water.Interface, doRecv func([]byte) error) {
	libol.Info("PointAuth.GoRecv: %s", dev.Name())

	defer dev.Close()
	for {
		data := make([]byte, w.ifMtu)
		n, err := dev.Read(data)
		if err != nil {
			libol.Error("PointAuth.GoRev: %s", err)
			break
		}

		libol.Debug("PointAuth.GoRev: % x\n", data[:n])
		if err := doRecv(data[:n]); err != nil {
			libol.Error("PointAuth.GoRev: do-recv %s %s", dev.Name(), err)
		}
	}
}

type WithRequest struct {
	worker *Worker
}

func NewWithRequest(worker *Worker, c *Config) (w *WithRequest) {
	w = &WithRequest{
		worker: worker,
	}
	return
}

func (w *WithRequest) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("WithRequest.OnFrame % x.", frame.Data)

	if libol.IsInst(frame.Data) {
		action, body := libol.DecActionBody(frame.Data)
		libol.Debug("WithRequest.OnFrame.action: %s %s", action, body)

		if action == "neig=" {
			//TODO
		}
	}

	return nil
}
