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
	Neighbor    *Neighborer
	Redis       *libol.RedisCli
	EnableRedis bool
	Conf        *Config

	br          tenus.Bridger
	brip        net.IP
	brnet       *net.IPNet
	keys        []int
	hooks       map[int]func(*libol.TcpClient, *libol.Frame) error
	ifmtu       int
	clientsLock sync.RWMutex
	clients     map[*libol.TcpClient]*Point
	usersLock   sync.RWMutex
	users       map[string]*User
	newtime     int64
	brname      string
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
		ifmtu:       c.Ifmtu,
		hooks:       make(map[int]func(*libol.TcpClient, *libol.Frame) error),
		keys:        make([]int, 0, 1024),
		clients:     make(map[*libol.TcpClient]*Point, 1024),
		users:       make(map[string]*User, 1024),
		newtime:     time.Now().Unix(),
		brname:      c.Brname,
		links:       make(map[string]*point.Point),
	}

	if err := w.Redis.Open(); err != nil {
		libol.Error("NewWorker: redis.Open %s", err)
	}
	w.Auth = NewPointAuth(w, c)
	w.Request = NewWithRequest(w, c)
	w.Neighbor = NewNeighborer(w, c)
	w.NewBr()
	w.Register()
	w.LoadUsers()
	w.LoadLinks()

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
	if w.brname == "" {
		addrs := strings.Split(w.Server.Addr, ":")
		if len(addrs) != 2 {
			w.brname = "brol-default"
		} else {
			w.brname = fmt.Sprintf("brol-%s", addrs[1])
		}
	}

	return w.brname
}

func (w *Worker) NewBr() {
	var err error
	var br tenus.Bridger

	addr := w.Conf.Ifaddr
	brname := w.BrName()
	br, err = tenus.BridgeFromName(brname)
	if err != nil {
		br, err = tenus.NewBridgeWithName(brname)
		if err != nil {
			libol.Error("Worker.newBr: %s", err)
		}
	}

	brctl := libol.NewBrCtl(brname)
	if err := brctl.Stp(true); err != nil {
		libol.Error("Worker.newBr.Stp: %s", err)
	}

	if err = br.SetLinkUp(); err != nil {
		libol.Error("Worker.newBr: %s", err)
	}

	libol.Info("Worker.newBr %s", brname)

	if addr != "" {
		ip, net, err := net.ParseCIDR(addr)
		if err != nil {
			libol.Error("Worker.newBr.ParseCIDR %s : %s", addr, err)
		}
		if err := br.SetLinkIp(ip, net); err != nil {
			libol.Error("Worker.newBr.SetLinkIp %s : %s", brname, err)
		}

		w.brip = ip
		w.brnet = net
	}

	w.br = br
}

func (w *Worker) NewTap() (*water.Interface, error) {
	libol.Debug("Worker.newTap")
	ifce, err := water.New(water.Config{
		DeviceType: water.TAP,
	})
	if err != nil {
		libol.Error("Worker.newTap: %s", err)
		return nil, err
	}

	link, err := tenus.NewLinkFrom(ifce.Name())
	if err != nil {
		libol.Error("Worker.newTap: Get ifce %s: %s", ifce.Name(), err)
		return nil, err
	}

	if err := link.SetLinkUp(); err != nil {
		libol.Error("Worker.newTap: ", err)
	}

	if err := w.br.AddSlaveIfc(link.NetInterface()); err != nil {
		libol.Error("Worker.newTap: Switch ifce %s: %s", ifce.Name(), err)
		return nil, err
	}

	libol.Info("Worker.newTap %s", ifce.Name())

	return ifce, nil
}

func (w *Worker) Start() {
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
	client.Status = libol.CL_CONNECTED

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

	ifce := point.Device
	if point == nil || point.Device == nil {
		return libol.Errer("Tap devices is nil")
	}

	if _, err := ifce.Write(data); err != nil {
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

func (w *Worker) Close() {
	libol.Info("Worker.Close")

	w.Server.Close()

	if w.br != nil && w.brip != nil {
		if err := w.br.UnsetLinkIp(w.brip, w.brnet); err != nil {
			libol.Error("Worker.Close.UnsetLinkIp %s : %s", w.br.NetInterface().Name, err)
		}
	}

	for _, p := range w.links {
		p.Close()
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

func (w *Worker) ListUser() chan *User {
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

func (w *Worker) ListPoint() chan *Point {
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
	return time.Now().Unix() - w.newtime
}

func (w *Worker) PubPoint(p *Point, isadd bool) {
	if !w.EnableRedis {
		return
	}

	key := fmt.Sprintf("point:%s", strings.Replace(p.Client.String(), ":", "-", -1))
	value := map[string]interface{}{
		"remote":  p.Client.String(),
		"newtime": p.Client.NewTime,
		"device":  p.Device.Name(),
		"actived": isadd,
	}

	if err := w.Redis.HMSet(key, value); err != nil {
		libol.Error("Worker.PubPoint hset %s", err)
	}
}

func (w *Worker) AddLink(c *point.Config) {
	c.Brname = w.BrName() //Reset bridge name.

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
		p.Close()
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

func (w *Worker) ListLink() chan *point.Point {
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
	ifmtu  int
	wroker *Worker
}

func NewPointAuth(wroker *Worker, c *Config) (w *PointAuth) {
	w = &PointAuth{
		ifmtu:  c.Ifmtu,
		wroker: wroker,
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

	if client.Status != libol.CL_AUTHED {
		client.Droped++
		libol.Debug("PointAuth.onRecv: %s unauth", client.Addr)
		return libol.Errer("Unauthed client.")
	}

	return nil
}

func (w *PointAuth) handleLogin(client *libol.TcpClient, data string) error {
	libol.Debug("PointAuth.handleLogin: %s", data)

	if client.Status == libol.CL_AUTHED {
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
	_user := w.wroker.GetUser(name)
	if _user != nil {
		if _user.Password == user.Password {
			client.Status = libol.CL_AUTHED
			libol.Info("PointAuth.handleLogin: %s Authed", client.Addr)
			w.onAuth(client)
			return nil
		}

		client.Status = libol.CL_UNAUTH
	}

	return libol.Errer("Auth failed.")
}

func (w *PointAuth) onAuth(client *libol.TcpClient) error {
	if client.Status != libol.CL_AUTHED {
		return libol.Errer("not authed.")
	}

	libol.Info("PointAuth.onAuth: %s", client.Addr)
	ifce, err := w.wroker.NewTap()
	if err != nil {
		return err
	}

	w.wroker.AddPoint(NewPoint(client, ifce))

	go w.GoRecv(ifce, client.SendMsg)

	return nil
}

func (w *PointAuth) GoRecv(ifce *water.Interface, doRecv func([]byte) error) {
	libol.Info("PointAuth.GoRecv: %s", ifce.Name())

	defer ifce.Close()
	for {
		data := make([]byte, w.ifmtu)
		n, err := ifce.Read(data)
		if err != nil {
			libol.Error("PointAuth.GoRev: %s", err)
			break
		}

		libol.Debug("PointAuth.GoRev: % x\n", data[:n])
		if err := doRecv(data[:n]); err != nil {
			libol.Error("PointAuth.GoRev: do-recv %s %s", ifce.Name(), err)
		}
	}
}

type WithRequest struct {
	wroker *Worker
}

func NewWithRequest(wroker *Worker, c *Config) (w *WithRequest) {
	w = &WithRequest{
		wroker: wroker,
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
