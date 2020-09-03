package olap

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olap/http"
	"net"
	"strings"
	"sync"
	"time"
)

type KeepAlive struct {
	Interval int64
	LastTime int64
}

func (k *KeepAlive) Should() bool {
	return time.Now().Unix()-k.LastTime >= k.Interval
}

func (k *KeepAlive) Update() {
	k.LastTime = time.Now().Unix()
}

type SocketWorkerListener struct {
	OnClose   func(w *SocketWorker) error
	OnSuccess func(w *SocketWorker) error
	OnIpAddr  func(w *SocketWorker, n *models.Network) error
	ReadAt    func(frame *libol.FrameMessage) error
}

var (
	EventConed   = "conned"
	EventRecon   = "reconn"
	EventClosed  = "closed"
	EventSuccess = "success"
	EventSignIn  = "signIn"
	EventLogin   = "login"
)

type SocketEvent struct {
	Type   string
	Reason string
	Time   int64
}

func (e SocketEvent) String() string {
	return e.Type + " " + e.Reason
}

func NewEvent(typ, reason string) SocketEvent {
	return SocketEvent{
		Time:   time.Now().Unix(),
		Reason: reason,
		Type:   typ,
	}
}

type jobTimer struct {
	Time int64
	Call func() error
}

const (
	rtLast      = "last"      // record time last frame received or connected.
	rtConnected = "connected" // record last connected time.
	rtReconnect = "reconnect" // record time when triggered reconnected.
	rtSuccess   = "reSuccess" // record success time when login.
	rtSleeps    = "sleeps"    // record times to control connecting delay.
	rtClosed    = "closed"
	rtLive      = "live" // record received pong frame time.
)

type SocketWorker struct {
	// private
	listener   SocketWorkerListener
	client     libol.SocketClient
	lock       sync.Mutex
	user       *models.User
	network    *models.Network
	routes     map[string]*models.Route
	keepalive  KeepAlive
	done       chan bool
	ticker     *time.Ticker
	pinCfg     *config.Point
	eventQueue chan SocketEvent
	writeQueue chan *libol.FrameMessage
	jobber     []jobTimer
	record     *libol.SafeStrInt64
	out        *libol.SubLogger
}

func NewSocketWorker(client libol.SocketClient, c *config.Point) *SocketWorker {
	t := &SocketWorker{
		client:  client,
		user:    models.NewUser(c.Username, c.Password),
		network: models.NewNetwork(c.Network, c.Interface.Address),
		routes:  make(map[string]*models.Route, 64),
		record:  libol.NewSafeStrInt64(),
		done:    make(chan bool, 2),
		ticker:  time.NewTicker(2 * time.Second),
		keepalive: KeepAlive{
			Interval: 15,
			LastTime: time.Now().Unix(),
		},
		pinCfg:     c,
		eventQueue: make(chan SocketEvent, 32),
		writeQueue: make(chan *libol.FrameMessage, 1024),
		jobber:     make([]jobTimer, 0, 32),
		out:        libol.NewSubLogger(c.Id()),
	}
	t.user.Alias = c.Alias
	t.user.Network = c.Network
	return t
}

func (t *SocketWorker) sleepNow() int64 {
	sleeps := t.record.Get(rtSleeps)
	return sleeps * 5
}

func (t *SocketWorker) sleepIdle() int64 {
	sleeps := t.record.Get(rtSleeps)
	if sleeps < 20 {
		t.record.Add(rtSleeps, 1)
	}
	return t.sleepNow()
}

func (t *SocketWorker) Initialize() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.out.Info("SocketWorker.Initialize")
	t.client.SetMaxSize(t.pinCfg.Interface.IfMtu)
	t.client.SetListener(libol.ClientListener{
		OnConnected: func(client libol.SocketClient) error {
			t.record.Set(rtConnected, time.Now().Unix())
			t.eventQueue <- NewEvent(EventConed, "from socket")
			return nil
		},
		OnClose: func(client libol.SocketClient) error {
			t.record.Set(rtClosed, time.Now().Unix())
			t.eventQueue <- NewEvent(EventClosed, "from socket")
			return nil
		},
	})
	t.record.Set(rtLast, time.Now().Unix())
	t.record.Set(rtReconnect, time.Now().Unix())
}

func (t *SocketWorker) Start() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.out.Info("SocketWorker.Start")
	_ = t.connect()
	libol.Go(t.Loop)
}

func (t *SocketWorker) sendLeave(client libol.SocketClient) error {
	if client == nil {
		return libol.NewErr("client is nil")
	}
	data := struct {
		DateTime   int64  `json:"datetime"`
		UUID       string `json:"uuid"`
		Alias      string `json:"alias"`
		Connection string `json:"connection"`
		Address    string `json:"address"`
	}{
		DateTime:   time.Now().Unix(),
		UUID:       t.user.UUID,
		Alias:      t.user.Alias,
		Address:    t.client.LocalAddr(),
		Connection: t.client.RemoteAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.out.Cmd("SocketWorker.leave: left: %s", body)
	m := libol.NewControlFrame(libol.LeftReq, body)
	if err := client.WriteMsg(m); err != nil {
		return err
	}
	return nil
}

func (t *SocketWorker) leave() {
	t.out.Info("SocketWorker.leave")
	if err := t.sendLeave(t.client); err != nil {
		t.out.Error("SocketWorker.leave: %s", err)
	}
}

func (t *SocketWorker) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.out.Info("SocketWorker.Stop")
	t.leave()
	t.client.Terminal()
	t.done <- true
	t.client = nil
	t.ticker.Stop()
}

func (t *SocketWorker) close() {
	if t.client != nil {
		t.client.Close()
	}
}

func (t *SocketWorker) connect() error {
	t.out.Warn("SocketWorker.connect %s:%d", t.client, libol.ClInit)
	t.client.Close()
	s := t.client.Status()
	if s != libol.ClInit {
		t.out.Warn("SocketWorker.connect %s %d->%d", t.client, s, libol.ClInit)
		t.client.SetStatus(libol.ClInit)
	}
	if err := t.client.Connect(); err != nil {
		t.out.Error("SocketWorker.connect %s %s", t.client, err)
		return err
	}
	return nil
}

func (t *SocketWorker) reconnect() {
	if t.isStopped() {
		return
	}
	t.record.Set(rtReconnect, time.Now().Unix())
	t.jobber = append(t.jobber, jobTimer{
		Time: time.Now().Unix() + t.sleepIdle(),
		Call: func() error {
			t.out.Debug("SocketWorker.reconnect: on jobber")
			if t.record.Get(rtConnected) >= t.record.Get(rtReconnect) { // already connected after.
				t.out.Cmd("SocketWorker.reconnect: dissed by connected")
				return nil
			}
			if t.record.Get(rtLast) >= t.record.Get(rtReconnect) { // ignored immediately connect.
				t.out.Info("SocketWorker.reconnect: dissed by last")
				return nil
			}
			t.out.Info("SocketWorker.reconnect: %v", t.record.Data())
			return t.connect()
		},
	})
}

func (t *SocketWorker) reLogin() error {
	return t.toLogin(t.client)
}

func (t *SocketWorker) sendLogin(client libol.SocketClient) error {
	if client == nil {
		return libol.NewErr("client is nil")
	}
	body, err := json.Marshal(t.user)
	if err != nil {
		return err
	}
	t.out.Cmd("SocketWorker.toLogin: %s", body)
	m := libol.NewControlFrame(libol.LoginReq, body)
	if err := client.WriteMsg(m); err != nil {
		return err
	}
	return nil
}

// toLogin request
func (t *SocketWorker) toLogin(client libol.SocketClient) error {
	if err := t.sendLogin(client); err != nil {
		t.out.Error("SocketWorker.toLogin: %s", err)
		return err
	}
	return nil
}

func (t *SocketWorker) sendIpAddr(client libol.SocketClient) error {
	if client == nil {
		return libol.NewErr("client is nil")
	}
	body, err := json.Marshal(t.network)
	if err != nil {
		return err
	}
	t.out.Cmd("SocketWorker.toNetwork: %s", body)
	m := libol.NewControlFrame(libol.IpAddrReq, body)
	if err := client.WriteMsg(m); err != nil {
		return err
	}
	return nil
}

// network request
func (t *SocketWorker) toNetwork(client libol.SocketClient) error {
	if !t.pinCfg.RequestAddr && t.network.IfAddr == "" {
		t.out.Info("SocketWorker.toNetwork: notNeed")
		return nil
	}
	if err := t.sendIpAddr(client); err != nil {
		t.out.Error("SocketWorker.toNetwork: %s", err)
		return err
	}
	return nil
}

func (t *SocketWorker) onLogin(resp []byte) error {
	if strings.HasPrefix(string(resp), "okay") {
		t.client.SetStatus(libol.ClAuth)
		if t.listener.OnSuccess != nil {
			_ = t.listener.OnSuccess(t)
		}
		t.record.Set(rtSleeps, 0)
		t.record.Set(rtSuccess, time.Now().Unix())
		_ = t.toNetwork(t.client)
		t.eventQueue <- NewEvent(EventSuccess, "already success")
		t.out.Info("SocketWorker.onInstruct.toLogin: success")
	} else {
		t.client.SetStatus(libol.ClUnAuth)
		t.out.Error("SocketWorker.onInstruct.toLogin: %s", resp)
	}
	return nil
}

func (t *SocketWorker) onIpAddr(resp []byte) error {
	if !t.pinCfg.RequestAddr {
		t.out.Info("SocketWorker.onIpAddr: notAllowed")
		return nil
	}
	n := &models.Network{}
	if err := json.Unmarshal(resp, n); err != nil {
		return libol.NewErr("SocketWorker.onInstruct: invalid json data.")
	}
	t.network = n
	if t.listener.OnIpAddr != nil {
		_ = t.listener.OnIpAddr(t, n)
	}
	return nil
}

func (t *SocketWorker) onLeft(resp []byte) error {
	t.out.Info("SocketWorker.onLeft")
	t.out.Cmd("SocketWorker.onLeft: %s", resp)
	t.close()
	return nil
}

func (t *SocketWorker) onSignIn(resp []byte) error {
	t.out.Info("SocketWorker.onSignIn")
	t.out.Cmd("SocketWorker.onSignIn: %s", resp)
	t.eventQueue <- NewEvent(EventSignIn, "request from server")
	return nil
}

// handle instruct from virtual switch
func (t *SocketWorker) onInstruct(frame *libol.FrameMessage) error {
	if !frame.IsControl() {
		return nil
	}
	action, resp := frame.CmdAndParams()
	if libol.HasLog(libol.CMD) {
		t.out.Cmd("SocketWorker.onInstruct %s %s", action, resp)
	}
	switch action {
	case libol.LoginResp:
		return t.onLogin(resp)
	case libol.IpAddrResp:
		return t.onIpAddr(resp)
	case libol.PongResp:
		t.record.Set(rtLive, time.Now().Unix())
	case libol.SignReq:
		return t.onSignIn(resp)
	case libol.LeftReq:
		return t.onLeft(resp)
	default:
		t.out.Warn("SocketWorker.onInstruct: %s %s", action, resp)
	}
	return nil
}

func (t *SocketWorker) sendPing(client libol.SocketClient) error {
	if client == nil {
		return libol.NewErr("client is nil")
	}
	data := struct {
		DateTime   int64  `json:"datetime"`
		UUID       string `json:"uuid"`
		Alias      string `json:"alias"`
		Connection string `json:"connection"`
		Address    string `json:"address"`
	}{
		DateTime:   time.Now().Unix(),
		UUID:       t.user.UUID,
		Alias:      t.user.Alias,
		Address:    t.client.LocalAddr(),
		Connection: t.client.RemoteAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.out.Cmd("SocketWorker.sendPing: ping= %s", body)
	m := libol.NewControlFrame(libol.PingReq, body)
	if err := client.WriteMsg(m); err != nil {
		return err
	}
	return nil
}

func (t *SocketWorker) doKeepalive() error {
	if !t.keepalive.Should() {
		return nil
	}
	t.keepalive.Update()
	if t.client.Have(libol.ClAuth) {
		if err := t.sendPing(t.client); err != nil {
			t.out.Error("SocketWorker.doKeepalive: %s", err)
			return err
		}
	} else {
		if err := t.sendLogin(t.client); err != nil {
			t.out.Error("SocketWorker.doKeepalive: %s", err)
			return err
		}
	}
	return nil
}

func (t *SocketWorker) doJober() error {
	// travel jobber and execute expired.
	now := time.Now().Unix()
	newTimer := make([]jobTimer, 0, 32)
	for _, t := range t.jobber {
		if now >= t.Time {
			_ = t.Call()
		} else {
			newTimer = append(newTimer, t)
		}
	}
	t.jobber = newTimer
	t.out.Debug("SocketWorker.doJober %d", len(t.jobber))
	return nil
}

func (t *SocketWorker) doTicker() error {
	_ = t.doKeepalive()
	_ = t.doJober()
	return nil
}

func (t *SocketWorker) dispatch(ev SocketEvent) {
	t.out.Info("SocketWorker.dispatch %v", ev)
	switch ev.Type {
	case EventConed:
		if t.client != nil {
			libol.Go(t.Read)
			_ = t.toLogin(t.client)
		}
	case EventSuccess:
		_ = t.sendPing(t.client)
	case EventRecon:
		t.reconnect()
	case EventSignIn, EventLogin:
		_ = t.reLogin()
	}
}

func (t *SocketWorker) Loop() {
	for {
		select {
		case e := <-t.eventQueue:
			t.lock.Lock()
			t.dispatch(e)
			t.lock.Unlock()
		case d := <-t.writeQueue:
			_ = t.DoWrite(d)
		case <-t.done:
			return
		case c := <-t.ticker.C:
			t.out.Log("SocketWorker.Ticker: at %s", c)
			t.lock.Lock()
			_ = t.doTicker()
			t.lock.Unlock()
		}
	}
}

func (t *SocketWorker) isStopped() bool {
	return t.client == nil || t.client.Have(libol.ClTerminal)
}

func (t *SocketWorker) Read() {
	for {
		t.lock.Lock()
		if t.isStopped() || !t.client.IsOk() {
			t.out.Error("SocketWorker.Read: %v", t.client)
			t.lock.Unlock()
			break
		}
		t.lock.Unlock()
		data, err := t.client.ReadMsg()
		t.lock.Lock()
		if err != nil {
			t.out.Error("SocketWorker.Read: %s", err)
			t.lock.Unlock()
			break
		}
		if t.out.Has(libol.DEBUG) {
			t.out.Debug("SocketWorker.Read: %x", data)
		}
		t.record.Set(rtLast, time.Now().Unix())
		if data.Size() <= 0 {
			t.lock.Unlock()
			continue
		}
		data.Decode()
		if data.IsControl() {
			_ = t.onInstruct(data)
			t.lock.Unlock()
			continue
		}
		t.lock.Unlock()
		if t.listener.ReadAt != nil {
			_ = t.listener.ReadAt(data)
		}
	}
	if !t.isStopped() {
		t.eventQueue <- NewEvent(EventRecon, "from read")
	}
}

func (t *SocketWorker) deadCheck() {
	out := int64(t.pinCfg.Timeout)
	now := time.Now().Unix()
	if now-t.record.Get(rtLast) < out {
		return
	}
	if now-t.record.Get(rtReconnect) < out { // timeout and avoid send reconn frequently.
		t.out.Cmd("SocketWorker.deadCheck: reconn frequently")
		return
	}
	t.eventQueue <- NewEvent(EventRecon, "from dead check")
}

func (t *SocketWorker) DoWrite(frame *libol.FrameMessage) error {
	if t.out.Has(libol.DEBUG) {
		t.out.Debug("SocketWorker.DoWrite: %x", frame)
	}
	t.lock.Lock()
	t.deadCheck()
	if t.client == nil {
		t.lock.Unlock()
		return libol.NewErr("client is nil")
	}
	if !t.client.Have(libol.ClAuth) {
		t.out.Debug("SocketWorker.DoWrite: dropping by unAuth")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()
	if err := t.client.WriteMsg(frame); err != nil {
		t.out.Error("SocketWorker.DoWrite: %s", err)
		t.eventQueue <- NewEvent(EventRecon, "from write")
		return err
	}
	return nil
}

func (t *SocketWorker) Auth() (string, string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.user.Name, t.user.Password
}

func (t *SocketWorker) SetAuth(auth string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	values := strings.Split(auth, ":")
	t.user.Name = values[0]
	if len(values) > 1 {
		t.user.Password = values[1]
	}
}

func (t *SocketWorker) SetUUID(v string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.user.UUID = v
}

type TapWorkerListener struct {
	OnOpen   func(w *TapWorker) error
	OnClose  func(w *TapWorker)
	FindNext func(dest []byte) []byte
	ReadAt   func(frame *libol.FrameMessage) error
}

type TunEther struct {
	HwAddr []byte
	IpAddr []byte
}

type TapWorker struct {
	// private
	lock       sync.Mutex
	device     network.Taper
	listener   TapWorkerListener
	ether      TunEther
	neighbor   Neighbors
	openAgain  bool
	deviceCfg  network.TapConfig
	pointCfg   *config.Point
	ifAddr     string
	writeQueue chan *libol.FrameMessage
	done       chan bool
	recvIpAddr chan string
	out        *libol.SubLogger
}

func NewTapWorker(devCfg network.TapConfig, c *config.Point) (a *TapWorker) {
	a = &TapWorker{
		device:     nil,
		deviceCfg:  devCfg,
		pointCfg:   c,
		openAgain:  false,
		done:       make(chan bool, 2),
		writeQueue: make(chan *libol.FrameMessage, 1024),
		recvIpAddr: make(chan string, 1024),
		out:        libol.NewSubLogger(c.Id()),
	}
	return
}

func (a *TapWorker) Initialize() {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.out.Info("TapWorker.Initialize")
	a.neighbor = Neighbors{
		neighbors: make(map[uint32]*Neighbor, 1024),
		done:      make(chan bool),
		ticker:    time.NewTicker(5 * time.Second),
		timeout:   3 * 60,
		interval:  60,
		listener: NeighborListener{
			Interval: func(dest []byte) {
				a.OnArpAlive(dest)
			},
			Expire: func(dest []byte) {
				a.OnArpAlive(dest)
			},
		},
	}
	a.open()
	if a.device != nil && a.device.IsTun() {
		a.setEther(a.pointCfg.Interface.Address)
		a.ether.HwAddr = libol.GenEthAddr(6)
		a.out.Info("TapWorker.Initialize: src %x", a.ether.HwAddr)
	}
}

func (a *TapWorker) setEther(addr string) {
	a.neighbor.Clear()
	// format ip address.
	addr = libol.IpAddrFormat(addr)
	ifAddr := strings.SplitN(addr, "/", 2)[0]
	a.ether.IpAddr = net.ParseIP(ifAddr).To4()
	if a.ether.IpAddr == nil {
		a.out.Warn("TapWorker.setEther: srcIp is nil")
		a.ether.IpAddr = []byte{0x00, 0x00, 0x00, 0x00}
	} else {
		a.out.Info("TapWorker.setEther: srcIp % x", a.ether.IpAddr)
	}
	// changed address need open device again.
	if a.ifAddr != "" && a.ifAddr != addr {
		a.out.Warn("TapWorker.setEther changed %s->%s", a.ifAddr, addr)
		a.openAgain = true
	}
	a.ifAddr = addr
}

func (a *TapWorker) OnIpAddr(addr string) {
	a.recvIpAddr <- addr
}

func (a *TapWorker) open() {
	if a.device != nil {
		a.close()
		if !a.openAgain {
			time.Sleep(5 * time.Second) // sleep 5s and release cpu.
		}
	}
	device, err := network.NewKernelTap(a.pointCfg.Network, a.deviceCfg)
	if err != nil {
		a.out.Error("TapWorker.open: %s", err)
		return
	}
	a.out.Info("TapWorker.open: >>> %s <<<", device.Name())
	a.device = device
	if a.listener.OnOpen != nil {
		_ = a.listener.OnOpen(a)
	}
}

func (a *TapWorker) newEth(t uint16, dst []byte) *libol.Ether {
	eth := libol.NewEther(t)
	eth.Dst = dst
	eth.Src = a.ether.HwAddr
	return eth
}

func (a *TapWorker) OnArpAlive(dest []byte) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.onMiss(dest)
}

// process if ethernet destination is missed
func (a *TapWorker) onMiss(dest []byte) {
	a.out.Debug("TapWorker.onMiss: %v.", dest)
	eth := a.newEth(libol.EthArp, libol.BROADED)
	reply := libol.NewArp()
	reply.OpCode = libol.ArpRequest
	reply.SIpAddr = a.ether.IpAddr
	reply.TIpAddr = dest
	reply.SHwAddr = a.ether.HwAddr
	reply.THwAddr = libol.ZEROED

	frame := libol.NewFrameMessage()
	frame.Append(eth.Encode())
	frame.Append(reply.Encode())
	a.out.Debug("TapWorker.onMiss: %x.", frame.Frame()[:64])
	if a.listener.ReadAt != nil {
		_ = a.listener.ReadAt(frame)
	}
}

func (a *TapWorker) onFrame(frame *libol.FrameMessage, data []byte) int {
	size := len(data)
	if a.device.IsTun() {
		iph, err := libol.NewIpv4FromFrame(data)
		if err != nil {
			a.out.Warn("TapWorker.onFrame: %s", err)
			return 0
		}
		dest := iph.Destination
		if a.listener.FindNext != nil {
			dest = a.listener.FindNext(dest)
		}
		neb := a.neighbor.GetByBytes(dest)
		if neb == nil {
			a.onMiss(dest)
			a.out.Debug("TapWorker.onFrame: onMiss neighbor %v", dest)
			return 0
		}
		eth := a.newEth(libol.EthIp4, neb.HwAddr)
		frame.Append(eth.Encode()) // insert ethernet header.
		size += eth.Len
	}
	frame.SetSize(size)
	return size
}

func (a *TapWorker) Read() {
	for {
		a.lock.Lock()
		if a.device == nil {
			a.lock.Unlock()
			break
		}
		frame := libol.NewFrameMessage()
		data := frame.Frame()
		if a.device.IsTun() {
			data = data[libol.EtherLen:]
		}
		a.lock.Unlock()
		n, err := a.device.Read(data)
		a.lock.Lock()
		if err != nil || a.openAgain || a.device == nil {
			a.out.Warn("TapWorker.Read: %s", err)
			a.open()
			a.openAgain = false // clear openAgain flags
			a.lock.Unlock()
			continue
		}
		if a.out.Has(libol.DEBUG) {
			a.out.Debug("TapWorker.Read: %x", data[:n])
		}
		if size := a.onFrame(frame, data[:n]); size == 0 {
			a.lock.Unlock()
			continue
		}
		a.lock.Unlock()
		if a.listener.ReadAt != nil {
			_ = a.listener.ReadAt(frame)
		}
	}
}

func (a *TapWorker) Loop() {
	for {
		select {
		case <-a.done:
			return
		case d := <-a.writeQueue:
			_ = a.DoWrite(d)
		case addr := <-a.recvIpAddr:
			a.lock.Lock()
			a.setEther(addr)
			a.lock.Unlock()
		}
	}
}

func (a *TapWorker) DoWrite(frame *libol.FrameMessage) error {
	data := frame.Frame()
	if a.out.Has(libol.DEBUG) {
		a.out.Debug("TapWorker.DoWrite: %x", data)
	}
	a.lock.Lock()
	if a.device == nil {
		a.lock.Unlock()
		return libol.NewErr("device is nil")
	}
	if a.device.IsTun() {
		// proxy arp request.
		if a.toArp(data) {
			a.out.Debug("TapWorker.DoWrite: Arp proxy.")
			a.lock.Unlock()
			return nil
		}
		eth, err := libol.NewEtherFromFrame(data)
		if err != nil {
			a.out.Error("TapWorker.DoWrite: %s", err)
			a.lock.Unlock()
			return nil
		}
		if eth.IsIP4() {
			data = data[14:]
		} else {
			a.out.Debug("TapWorker.DoWrite: 0x%04x not IPv4", eth.Type)
			a.lock.Unlock()
			return nil
		}
	}
	a.lock.Unlock()
	if _, err := a.device.Write(data); err != nil {
		a.out.Error("TapWorker.DoWrite: %s", err)
		return err
	}
	return nil
}

// learn source from arp
func (a *TapWorker) toArp(data []byte) bool {
	a.out.Debug("TapWorker.toArp")
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		a.out.Warn("TapWorker.toArp: %s", err)
		return false
	}
	if !eth.IsArp() {
		return false
	}
	arp, err := libol.NewArpFromFrame(data[eth.Len:])
	if err != nil {
		a.out.Error("TapWorker.toArp: %s.", err)
		return false
	}
	if arp.IsIP4() {
		if !bytes.Equal(eth.Src, arp.SHwAddr) {
			a.out.Error("TapWorker.toArp: eth.dst not arp.shw %x.", arp.SIpAddr)
			return true
		}
		switch arp.OpCode {
		case libol.ArpRequest:
			if bytes.Equal(arp.TIpAddr, a.ether.IpAddr) {
				eth := a.newEth(libol.EthArp, arp.SHwAddr)
				rep := libol.NewArp()
				rep.OpCode = libol.ArpReply
				rep.SIpAddr = a.ether.IpAddr
				rep.TIpAddr = arp.SIpAddr
				rep.SHwAddr = a.ether.HwAddr
				rep.THwAddr = arp.SHwAddr
				frame := libol.NewFrameMessage()
				frame.Append(eth.Encode())
				frame.Append(rep.Encode())
				a.out.Cmd1("TapWorker.toArp: reply %v on %x.", rep.SIpAddr, rep.SHwAddr)
				if a.listener.ReadAt != nil {
					_ = a.listener.ReadAt(frame)
				}
			}
		case libol.ArpReply:
			// TODO learn by request.
			if bytes.Equal(arp.THwAddr, a.ether.HwAddr) {
				a.neighbor.Add(&Neighbor{
					HwAddr:  arp.SHwAddr,
					IpAddr:  arp.SIpAddr,
					NewTime: time.Now().Unix(),
					Uptime:  time.Now().Unix(),
				})
				a.out.Cmd1("TapWorker.toArp: recv %v on %x.", arp.SIpAddr, arp.SHwAddr)
			}
		default:
			a.out.Warn("TapWorker.toArp: not op %x.", arp.OpCode)
		}
	}
	return true
}

func (a *TapWorker) close() {
	a.out.Info("TapWorker.close")
	if a.device != nil {
		if a.listener.OnClose != nil {
			a.listener.OnClose(a)
		}
		_ = a.device.Close()
		a.device = nil
	}
}

func (a *TapWorker) Start() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.out.Info("TapWorker.Start")
	libol.Go(a.Read)
	libol.Go(a.Loop)
	libol.Go(a.neighbor.Start)
}

func (a *TapWorker) Stop() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.out.Info("TapWorker.Stop")
	a.done <- true
	a.neighbor.Stop()
	a.close()
}

type WorkerListener struct {
	AddAddr   func(ipStr string) error
	DelAddr   func(ipStr string) error
	OnTap     func(w *TapWorker) error
	AddRoutes func(routes []*models.Route) error
	DelRoutes func(routes []*models.Route) error
}

type PrefixRule struct {
	Type        int
	Destination net.IPNet
	NextHop     net.IP
}

func GetSocketClient(c *config.Point) libol.SocketClient {
	switch c.Protocol {
	case "kcp":
		cfg := &libol.KcpConfig{
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewKcpClient(c.Connection, cfg)
	case "tcp":
		cfg := &libol.TcpConfig{
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewTcpClient(c.Connection, cfg)
	case "udp":
		cfg := &libol.UdpConfig{
			Block:   config.GetBlock(c.Crypt),
			Timeout: time.Duration(c.Timeout) * time.Second,
		}
		return libol.NewUdpClient(c.Connection, cfg)
	case "ws":
		cfg := &libol.WebConfig{
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewWebClient(c.Connection, cfg)
	case "wss":
		cfg := &libol.WebConfig{
			Ca:    &libol.WebCa{},
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewWebClient(c.Connection, cfg)
	default:
		cfg := &libol.TcpConfig{
			Tls:   &tls.Config{InsecureSkipVerify: true},
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewTcpClient(c.Connection, cfg)
	}
}

func GetTapCfg(c *config.Point) network.TapConfig {
	if c.Interface.Provider == "tun" {
		return network.TapConfig{
			Type:    network.TUN,
			Name:    c.Interface.Name,
			Network: c.Interface.Address,
		}
	} else {
		return network.TapConfig{
			Type:    network.TAP,
			Name:    c.Interface.Name,
			Network: c.Interface.Address,
		}
	}
}

type Worker struct {
	// private
	ifAddr    string
	listener  WorkerListener
	http      *http.Http
	conWorker *SocketWorker
	tapWorker *TapWorker
	cfg       *config.Point
	uuid      string
	network   *models.Network
	routes    []PrefixRule
	out       *libol.SubLogger
}

func NewWorker(cfg *config.Point) *Worker {
	return &Worker{
		ifAddr: cfg.Interface.Address,
		cfg:    cfg,
		routes: make([]PrefixRule, 0, 32),
		out:    libol.NewSubLogger(cfg.Id()),
	}
}

func (p *Worker) Initialize() {
	if p.cfg == nil {
		return
	}
	p.out.Info("Worker.Initialize")
	client := GetSocketClient(p.cfg)
	p.conWorker = NewSocketWorker(client, p.cfg)

	tapCfg := GetTapCfg(p.cfg)
	// register listener
	p.tapWorker = NewTapWorker(tapCfg, p.cfg)

	p.conWorker.SetUUID(p.UUID())
	p.conWorker.listener = SocketWorkerListener{
		OnClose:   p.OnClose,
		OnSuccess: p.OnSuccess,
		OnIpAddr:  p.OnIpAddr,
		ReadAt: func(frame *libol.FrameMessage) error {
			p.tapWorker.writeQueue <- frame
			return nil
		},
	}
	p.conWorker.Initialize()

	p.tapWorker.listener = TapWorkerListener{
		OnOpen: func(w *TapWorker) error {
			if p.listener.OnTap != nil {
				if err := p.listener.OnTap(w); err != nil {
					return err
				}
			}
			if p.network != nil {
				n := p.network
				// remove older firstly
				p.FreeIpAddr()
				_ = p.OnIpAddr(p.conWorker, n)
			}
			return nil
		},
		ReadAt: func(frame *libol.FrameMessage) error {
			p.conWorker.writeQueue <- frame
			return nil
		},
		FindNext: p.FindNext,
	}
	p.tapWorker.Initialize()

	if p.cfg.Http != nil {
		p.http = http.NewHttp(p)
	}
}

func (p *Worker) Start() {
	p.out.Debug("Worker.Start linux.")
	if p.cfg.PProf != "" {
		f := libol.PProf{Listen: p.cfg.PProf}
		f.Start()
	}
	p.tapWorker.Start()
	p.conWorker.Start()
	if p.http != nil {
		libol.Go(p.http.Start)
	}
}

func (p *Worker) Stop() {
	if p.tapWorker == nil || p.conWorker == nil {
		return
	}
	if p.http != nil {
		p.http.Shutdown()
	}
	p.FreeIpAddr()
	p.conWorker.Stop()
	p.tapWorker.Stop()
	p.conWorker = nil
	p.tapWorker = nil
}

func (p *Worker) Client() libol.SocketClient {
	if p.conWorker != nil {
		return p.conWorker.client
	}
	return nil
}

func (p *Worker) Device() network.Taper {
	if p.tapWorker != nil {
		return p.tapWorker.device
	}
	return nil
}

func (p *Worker) UpTime() int64 {
	client := p.Client()
	if client != nil {
		return client.AliveTime()
	}
	return 0
}

func (p *Worker) State() string {
	client := p.Client()
	if client != nil {
		return client.State()
	}
	return ""
}

func (p *Worker) Addr() string {
	client := p.Client()
	if client != nil {
		return client.Address()
	}
	return ""
}

func (p *Worker) IfName() string {
	dev := p.Device()
	if dev != nil {
		return dev.Name()
	}
	return ""
}

func (p *Worker) Worker() *SocketWorker {
	if p.conWorker != nil {
		return p.conWorker
	}
	return nil
}

func (p *Worker) FindNext(dest []byte) []byte {
	for _, rt := range p.routes {
		if !rt.Destination.Contains(dest) {
			continue
		}
		if rt.Type == 0x00 {
			break
		}
		if p.out.Has(libol.DEBUG) {
			p.out.Debug("Worker.FindNext %v to %v", dest, rt.NextHop)
		}
		return rt.NextHop.To4()
	}
	return dest
}

func (p *Worker) OnIpAddr(w *SocketWorker, n *models.Network) error {
	addr := fmt.Sprintf("%s/%s", n.IfAddr, n.Netmask)
	if models.NetworkEqual(p.network, n) {
		p.out.Debug("Worker.OnIpAddr: %s noChanged", addr)
		return nil
	}
	p.out.Cmd("Worker.OnIpAddr: %s", addr)
	p.out.Cmd("Worker.OnIpAddr: %s", n.Routes)
	prefix := libol.Netmask2Len(n.Netmask)
	ipStr := fmt.Sprintf("%s/%d", n.IfAddr, prefix)
	p.tapWorker.OnIpAddr(ipStr)
	if p.listener.AddAddr != nil {
		_ = p.listener.AddAddr(ipStr)
	}
	if p.listener.AddRoutes != nil {
		_ = p.listener.AddRoutes(n.Routes)
	}
	p.network = n
	// update routes
	ip := net.ParseIP(p.network.IfAddr)
	m := net.IPMask(net.ParseIP(p.network.Netmask).To4())
	p.routes = append(p.routes, PrefixRule{
		Type:        0x00,
		Destination: net.IPNet{IP: ip.Mask(m), Mask: m},
		NextHop:     libol.ZEROED,
	})
	for _, rt := range n.Routes {
		_, dest, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nxt := net.ParseIP(rt.NextHop)
		p.routes = append(p.routes, PrefixRule{
			Type:        0x01,
			Destination: *dest,
			NextHop:     nxt,
		})
	}
	return nil
}

func (p *Worker) FreeIpAddr() {
	if p.network == nil {
		return
	}
	if p.listener.DelRoutes != nil {
		_ = p.listener.DelRoutes(p.network.Routes)
	}
	if p.listener.DelAddr != nil {
		prefix := libol.Netmask2Len(p.network.Netmask)
		ipStr := fmt.Sprintf("%s/%d", p.network.IfAddr, prefix)
		_ = p.listener.DelAddr(ipStr)
	}
	p.network = nil
	p.routes = make([]PrefixRule, 0, 32)
}

func (p *Worker) OnClose(w *SocketWorker) error {
	p.out.Info("Worker.OnClose")
	p.FreeIpAddr()
	return nil
}

func (p *Worker) OnSuccess(w *SocketWorker) error {
	p.out.Info("Worker.OnSuccess")
	if p.listener.AddAddr != nil {
		_ = p.listener.AddAddr(p.ifAddr)
	}
	return nil
}

func (p *Worker) UUID() string {
	if p.uuid == "" {
		p.uuid = libol.GenToken(32)
	}
	return p.uuid
}

func (p *Worker) SetUUID(v string) {
	p.uuid = v
}

func (p *Worker) Config() *config.Point {
	return p.cfg
}
