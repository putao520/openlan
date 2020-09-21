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
	"net"
	"runtime"
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

var (
	EvSocConed   = "conned"
	EvSocRecon   = "reconn"
	EvSocClosed  = "closed"
	EvSocSuccess = "success"
	EvSocSignIn  = "signIn"
	EvSocLogin   = "login"
	EvTapIpAddr  = "ipAddr"
	EvTapReadErr = "readErr"
	EvTapReset   = "reset"
	EvTapOpenErr = "openErr"
)

type WorkerEvent struct {
	Type   string
	Reason string
	Time   int64
	Data   interface{}
}

func (e *WorkerEvent) String() string {
	return e.Type + " " + e.Reason
}

func NewEvent(newType, reason string) *WorkerEvent {
	return &WorkerEvent{
		Type:   newType,
		Time:   time.Now().Unix(),
		Reason: reason,
	}
}

type SocketWorkerListener struct {
	OnClose   func(w *SocketWorker) error
	OnSuccess func(w *SocketWorker) error
	OnIpAddr  func(w *SocketWorker, n *models.Network) error
	ReadAt    func(frame *libol.FrameMessage) error
}

type jobTimer struct {
	Time int64
	Call func() error
}

const (
	rtLast      = "lastAt"   // record time last frame received or connected.
	rtConnected = "connAt"   // record last connected time.
	rtReConnect = "reconnAt" // record time when triggered reconnected.
	rtSuccess   = "succAt"   // record success time when login.
	rtSleeps    = "sleeps"   // record times to control connecting delay.
	rtClosed    = "clsAt"    // close time
	rtLive      = "liveAt"   // record received pong frame time.
	rtIpAddr    = "addrAt"   // record last receive ipAddr message after success.
	rtConnects  = "conns"    // record times of reconnecting
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
	eventQueue chan *WorkerEvent
	writeQueue chan *libol.FrameMessage
	jobber     []jobTimer
	record     *libol.SafeStrInt64
	out        *libol.SubLogger
	wlFrame    *libol.FrameMessage // Last frame from write.
}

func NewSocketWorker(client libol.SocketClient, c *config.Point) *SocketWorker {
	t := &SocketWorker{
		client:     client,
		network:    models.NewNetwork(c.Network, c.Interface.Address),
		routes:     make(map[string]*models.Route, 64),
		record:     libol.NewSafeStrInt64(),
		done:       make(chan bool, 2),
		ticker:     time.NewTicker(2 * time.Second),
		pinCfg:     c,
		eventQueue: make(chan *WorkerEvent, 32),
		writeQueue: make(chan *libol.FrameMessage, c.Queue.SockWr),
		jobber:     make([]jobTimer, 0, 32),
		out:        libol.NewSubLogger(c.Id()),
	}
	t.user = &models.User{
		Alias:    c.Alias,
		Name:     c.Username,
		Password: c.Password,
		Network:  c.Network,
		System:   runtime.GOOS,
	}
	t.keepalive = KeepAlive{
		Interval: 15,
		LastTime: time.Now().Unix(),
	}
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
			t.eventQueue <- NewEvent(EvSocConed, "from socket")
			return nil
		},
		OnClose: func(client libol.SocketClient) error {
			t.record.Set(rtClosed, time.Now().Unix())
			t.eventQueue <- NewEvent(EvSocClosed, "from socket")
			return nil
		},
	})
	t.record.Set(rtLast, time.Now().Unix())
	t.record.Set(rtReConnect, time.Now().Unix())
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
	t.out.Warn("SocketWorker.connect: %s", t.client)
	t.client.Close()
	s := t.client.Status()
	if s != libol.ClInit {
		t.out.Warn("SocketWorker.connect: %s %s", t.client, s)
		t.client.SetStatus(libol.ClInit)
	}
	t.record.Add(rtConnects, 1)
	if err := t.client.Connect(); err != nil {
		t.out.Error("SocketWorker.connect: %s %s", t.client, err)
		return err
	}
	return nil
}

func (t *SocketWorker) reconnect() {
	if t.isStopped() {
		return
	}
	t.record.Set(rtReConnect, time.Now().Unix())
	job := jobTimer{
		Time: time.Now().Unix() + t.sleepIdle(),
		Call: func() error {
			t.out.Debug("SocketWorker.reconnect: on jobber")
			rtConn := t.record.Get(rtConnected)
			rtReCon := t.record.Get(rtReConnect)
			rtLast := t.record.Get(rtLast)
			rtLive := t.record.Get(rtLive)
			if rtConn >= rtReCon { // already connected after.
				t.out.Cmd("SocketWorker.reconnect: dissed by connected")
				return nil
			}
			t.out.Info("SocketWorker.reconnect: l: %d a: %d", rtLast, rtLive)
			t.out.Info("SocketWorker.reconnect: c: %d r: %d", rtConn, rtReCon)
			return t.connect()
		},
	}
	t.jobber = append(t.jobber, job)
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

func (t *SocketWorker) canReqAddr() bool {
	if t.pinCfg.RequestAddr {
		return true
	}
	// For link, need advise ipAddr with configured address.
	if t.network.IfAddr != "" {
		return true
	}
	return false
}

// network request
func (t *SocketWorker) toNetwork(client libol.SocketClient) error {
	if !t.canReqAddr() {
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
	if t.client.Have(libol.ClAuth) {
		t.out.Cmd("SocketWorker.onLogin: %s", resp)
		return nil
	}
	if strings.HasPrefix(string(resp), "okay") {
		t.client.SetStatus(libol.ClAuth)
		if t.listener.OnSuccess != nil {
			_ = t.listener.OnSuccess(t)
		}
		t.record.Set(rtSleeps, 0)
		t.record.Set(rtIpAddr, 0)
		t.record.Set(rtSuccess, time.Now().Unix())
		t.eventQueue <- NewEvent(EvSocSuccess, "from login")
		t.out.Info("SocketWorker.onLogin: success")
	} else {
		t.client.SetStatus(libol.ClUnAuth)
		t.out.Error("SocketWorker.onLogin: %s", resp)
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
		return libol.NewErr("SocketWorker.onIpAddr: invalid json data.")
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
	t.eventQueue <- NewEvent(EvSocSignIn, "request from server")
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
		t.record.Set(rtIpAddr, time.Now().Unix())
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

func (t *SocketWorker) keepAlive() {
	if !t.keepalive.Should() {
		return
	}
	t.keepalive.Update()
	if t.client.Have(libol.ClAuth) {
		// Whether ipAddr request was already? and try ipAddr?
		rtIp := t.record.Get(rtIpAddr)
		rtSuc := t.record.Get(rtSuccess)
		if t.canReqAddr() && rtIp < rtSuc {
			_ = t.toNetwork(t.client)
		}
		if err := t.sendPing(t.client); err != nil {
			t.out.Error("SocketWorker.keepAlive: %s", err)
		}
	} else {
		if err := t.sendLogin(t.client); err != nil {
			t.out.Error("SocketWorker.keepAlive: %s", err)
		}
	}
}

func (t *SocketWorker) checkJobber() {
	// travel jobber and execute it expired.
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
	t.out.Debug("SocketWorker.checkJobber: %d", len(t.jobber))
}

func (t *SocketWorker) checkAlive() {
	out := int64(t.pinCfg.Timeout)
	now := time.Now().Unix()
	if now-t.record.Get(rtLast) < out || now-t.record.Get(rtLive) < out {
		return
	}
	if now-t.record.Get(rtReConnect) < out { // timeout and avoid send reconn frequently.
		t.out.Cmd("SocketWorker.checkAlive: reconn frequently")
		return
	}
	t.eventQueue <- NewEvent(EvSocRecon, "from alive check")
}

func (t *SocketWorker) doTicker() error {
	t.checkAlive()  // period to check whether alive.
	t.keepAlive()   // send ping and wait pong to keep alive.
	t.checkJobber() // period to check job whether timeout.
	return nil
}

func (t *SocketWorker) dispatch(ev *WorkerEvent) {
	t.out.Event("SocketWorker.dispatch: %v", ev)
	switch ev.Type {
	case EvSocConed:
		if t.client != nil {
			_ = t.toLogin(t.client)
			libol.Go(func() {
				t.Read(t.client)
			})
		}
	case EvSocSuccess:
		_ = t.toNetwork(t.client)
	case EvSocRecon:
		t.out.Info("SocketWorker.dispatch: %v", ev)
		t.reconnect()
	case EvSocSignIn, EvSocLogin:
		_ = t.toLogin(t.client)
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

func (t *SocketWorker) Read(client libol.SocketClient) {
	for {
		data, err := client.ReadMsg()
		if err != nil {
			t.out.Error("SocketWorker.Read: %s", err)
			client.Close()
			break
		}
		if t.out.Has(libol.DEBUG) {
			t.out.Debug("SocketWorker.Read: %x", data)
		}
		if data.Size() <= 0 {
			continue
		}
		data.Decode()
		if data.IsControl() {
			t.lock.Lock()
			_ = t.onInstruct(data)
			t.lock.Unlock()
			continue
		}
		t.record.Set(rtLast, time.Now().Unix())
		if t.listener.ReadAt != nil {
			_ = t.listener.ReadAt(data)
		}
	}
	if !t.isStopped() {
		t.eventQueue <- NewEvent(EvSocRecon, "from read")
	}
}

func (t *SocketWorker) DoWrite(frame *libol.FrameMessage) error {
	if t.out.Has(libol.DEBUG) {
		t.out.Debug("SocketWorker.DoWrite: %x", frame)
	}
	t.checkAlive() // alive check immediately
	t.lock.Lock()
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
		t.out.Debug("SocketWorker.DoWrite: %s", err)
		return err
	}
	return nil
}

func (t *SocketWorker) Write(frame *libol.FrameMessage) error {
	t.writeQueue <- frame
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
	devCfg     network.TapConfig
	pinCfg     *config.Point
	ifAddr     string
	writeQueue chan *libol.FrameMessage
	done       chan bool
	out        *libol.SubLogger
	eventQueue chan *WorkerEvent
}

func NewTapWorker(devCfg network.TapConfig, pinCfg *config.Point) (a *TapWorker) {
	a = &TapWorker{
		devCfg:     devCfg,
		pinCfg:     pinCfg,
		done:       make(chan bool, 2),
		writeQueue: make(chan *libol.FrameMessage, pinCfg.Queue.TapWr),
		out:        libol.NewSubLogger(pinCfg.Id()),
		eventQueue: make(chan *WorkerEvent, 32),
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
	if a.IsTun() {
		addr := a.pinCfg.Interface.Address
		a.setEther(addr, libol.GenEthAddr(6))
		a.out.Info("TapWorker.Initialize: src %x", a.ether.HwAddr)
	}
	if err := a.open(); err != nil {
		a.eventQueue <- NewEvent(EvTapOpenErr, err.Error())
	}
}

func (a *TapWorker) IsTun() bool {
	return a.devCfg.Type == network.TUN
}

func (a *TapWorker) setEther(ipAddr string, hwAddr []byte) {
	a.neighbor.Clear()
	// format ip address.
	ipAddr = libol.IpAddrFormat(ipAddr)
	ifAddr := strings.SplitN(ipAddr, "/", 2)[0]
	a.ether.IpAddr = net.ParseIP(ifAddr).To4()
	if a.ether.IpAddr == nil {
		a.ether.IpAddr = []byte{0x00, 0x00, 0x00, 0x00}
	}
	a.out.Info("TapWorker.setEther: srcIp % x", a.ether.IpAddr)
	if hwAddr != nil {
		a.ether.HwAddr = hwAddr
	}
	// changed address need open device again.
	if a.ifAddr != "" && a.ifAddr != ipAddr {
		a.out.Warn("TapWorker.setEther changed %s->%s", a.ifAddr, ipAddr)
		a.eventQueue <- NewEvent(EvTapReset, "ifAddr changed")
	}
	a.ifAddr = ipAddr
}

func (a *TapWorker) OnIpAddr(addr string) {
	a.eventQueue <- NewEvent(EvTapIpAddr, addr)
}

func (a *TapWorker) open() error {
	a.close()
	device, err := network.NewTaper(a.pinCfg.Network, a.devCfg)
	if err != nil {
		a.out.Error("TapWorker.open: %s", err)
		return err
	}
	device.Up() // up device firstly
	libol.Go(func() {
		a.Read(device)
	})
	a.out.Info("TapWorker.open: >>> %s <<<", device.Name())
	a.device = device
	if a.listener.OnOpen != nil {
		_ = a.listener.OnOpen(a)
	}
	return nil
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
	eth := a.newEth(libol.EthArp, libol.EthAll)
	reply := libol.NewArp()
	reply.OpCode = libol.ArpRequest
	reply.SIpAddr = a.ether.IpAddr
	reply.TIpAddr = dest
	reply.SHwAddr = a.ether.HwAddr
	reply.THwAddr = libol.EthZero

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
	if a.IsTun() {
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

func (a *TapWorker) Read(device network.Taper) {
	for {
		frame := libol.NewFrameMessage()
		data := frame.Frame()
		if a.IsTun() {
			data = data[libol.EtherLen:]
		}
		if n, err := device.Read(data); err != nil {
			a.out.Error("TapWorker.Read: %s", err)
			break
		} else {
			if a.out.Has(libol.DEBUG) {
				a.out.Debug("TapWorker.Read: %x", data[:n])
			}
			if size := a.onFrame(frame, data[:n]); size == 0 {
				continue
			}
			if a.listener.ReadAt != nil {
				_ = a.listener.ReadAt(frame)
			}
		}
	}
	if !a.isStopped() {
		a.eventQueue <- NewEvent(EvTapReadErr, "from read")
	}
}

func (a *TapWorker) dispatch(ev *WorkerEvent) {
	a.out.Event("TapWorker.dispatch: %s", ev)
	switch ev.Type {
	case EvTapReadErr, EvTapOpenErr, EvTapReset:
		if err := a.open(); err != nil {
			time.Sleep(time.Second * 2)
			a.eventQueue <- NewEvent(EvTapOpenErr, err.Error())
		}
	case EvTapIpAddr:
		a.setEther(ev.Reason, nil)
	}
}

func (a *TapWorker) Loop() {
	for {
		select {
		case <-a.done:
			return
		case d := <-a.writeQueue:
			_ = a.DoWrite(d)
		case ev := <-a.eventQueue:
			a.lock.Lock()
			a.dispatch(ev)
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

func (a *TapWorker) Write(frame *libol.FrameMessage) error {
	a.writeQueue <- frame
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
				a.out.Event("TapWorker.toArp: reply %v on %x.", rep.SIpAddr, rep.SHwAddr)
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
				a.out.Event("TapWorker.toArp: recv %v on %x.", arp.SIpAddr, arp.SHwAddr)
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
	}
}

func (a *TapWorker) Start() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.out.Info("TapWorker.Start")
	libol.Go(a.Loop)
	libol.Go(a.neighbor.Start)
}

func (a *TapWorker) isStopped() bool {
	return a.device == nil
}

func (a *TapWorker) Stop() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.out.Info("TapWorker.Stop")
	a.done <- true
	a.neighbor.Stop()
	a.close()
	a.device = nil
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

func GetSocketClient(p *config.Point) libol.SocketClient {
	switch p.Protocol {
	case "kcp":
		c := &libol.KcpConfig{
			Block: config.GetBlock(p.Crypt),
		}
		return libol.NewKcpClient(p.Connection, c)
	case "tcp":
		c := &libol.TcpConfig{
			Block: config.GetBlock(p.Crypt),
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		return libol.NewTcpClient(p.Connection, c)
	case "udp":
		c := &libol.UdpConfig{
			Block:   config.GetBlock(p.Crypt),
			Timeout: time.Duration(p.Timeout) * time.Second,
		}
		return libol.NewUdpClient(p.Connection, c)
	case "ws":
		c := &libol.WebConfig{
			Block: config.GetBlock(p.Crypt),
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		return libol.NewWebClient(p.Connection, c)
	case "wss":
		c := &libol.WebConfig{
			Block: config.GetBlock(p.Crypt),
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		if p.Cert != nil {
			c.Cert = &libol.WebCert{
				Insecure: p.Cert.Insecure,
				RootCa:   p.Cert.CaFile,
			}
		}
		return libol.NewWebClient(p.Connection, c)
	default:
		c := &libol.TcpConfig{
			Block: config.GetBlock(p.Crypt),
			RdQus: p.Queue.SockRd,
			WrQus: p.Queue.SockWr,
		}
		if p.Cert != nil {
			c.Tls = &tls.Config{
				InsecureSkipVerify: p.Cert.Insecure,
				RootCAs:            p.Cert.GetCertPool(),
			}
		}
		return libol.NewTcpClient(p.Connection, c)
	}
}

func GetTapCfg(c *config.Point) network.TapConfig {
	cfg := network.TapConfig{
		Provider: c.Interface.Provider,
		Name:     c.Interface.Name,
		Network:  c.Interface.Address,
		KernBuf:  c.Queue.VirSnd,
		VirtBuf:  c.Queue.VirWrt,
	}
	if c.Interface.Provider == "tun" {
		cfg.Type = network.TUN
	} else {
		cfg.Type = network.TAP
	}
	return cfg
}

type Worker struct {
	// private
	ifAddr    string
	listener  WorkerListener
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

func (w *Worker) Initialize() {
	if w.cfg == nil {
		return
	}
	w.out.Info("Worker.Initialize")
	client := GetSocketClient(w.cfg)
	w.conWorker = NewSocketWorker(client, w.cfg)

	tapCfg := GetTapCfg(w.cfg)
	// register listener
	w.tapWorker = NewTapWorker(tapCfg, w.cfg)

	w.conWorker.SetUUID(w.UUID())
	w.conWorker.listener = SocketWorkerListener{
		OnClose:   w.OnClose,
		OnSuccess: w.OnSuccess,
		OnIpAddr:  w.OnIpAddr,
		ReadAt:    w.tapWorker.Write,
	}
	w.conWorker.Initialize()

	w.tapWorker.listener = TapWorkerListener{
		OnOpen: func(t *TapWorker) error {
			if w.listener.OnTap != nil {
				if err := w.listener.OnTap(t); err != nil {
					return err
				}
			}
			if w.network != nil {
				n := w.network
				// remove older firstly
				w.FreeIpAddr()
				_ = w.OnIpAddr(w.conWorker, n)
			}
			return nil
		},
		ReadAt:   w.conWorker.Write,
		FindNext: w.FindNext,
	}
	w.tapWorker.Initialize()
}

func (w *Worker) Start() {
	w.out.Debug("Worker.Start linux.")
	w.tapWorker.Start()
	w.conWorker.Start()
}

func (w *Worker) Stop() {
	if w.tapWorker == nil || w.conWorker == nil {
		return
	}
	w.FreeIpAddr()
	w.conWorker.Stop()
	w.tapWorker.Stop()
	w.conWorker = nil
	w.tapWorker = nil
}

func (w *Worker) UpTime() int64 {
	client := w.conWorker.client
	if client != nil {
		return client.AliveTime()
	}
	return 0
}

func (w *Worker) FindNext(dest []byte) []byte {
	for _, rt := range w.routes {
		if !rt.Destination.Contains(dest) {
			continue
		}
		if rt.Type == 0x00 {
			break
		}
		if w.out.Has(libol.DEBUG) {
			w.out.Debug("Worker.FindNext %v to %v", dest, rt.NextHop)
		}
		return rt.NextHop.To4()
	}
	return dest
}

func (w *Worker) OnIpAddr(s *SocketWorker, n *models.Network) error {
	addr := fmt.Sprintf("%s/%s", n.IfAddr, n.Netmask)
	if models.NetworkEqual(w.network, n) {
		w.out.Debug("Worker.OnIpAddr: %s noChanged", addr)
		return nil
	}
	w.out.Cmd("Worker.OnIpAddr: %s", addr)
	w.out.Cmd("Worker.OnIpAddr: %s", n.Routes)
	prefix := libol.Netmask2Len(n.Netmask)
	ipStr := fmt.Sprintf("%s/%d", n.IfAddr, prefix)
	w.tapWorker.OnIpAddr(ipStr)
	if w.listener.AddAddr != nil {
		_ = w.listener.AddAddr(ipStr)
	}
	if w.listener.AddRoutes != nil {
		_ = w.listener.AddRoutes(n.Routes)
	}
	w.network = n
	// update routes
	ip := net.ParseIP(w.network.IfAddr)
	m := net.IPMask(net.ParseIP(w.network.Netmask).To4())
	w.routes = append(w.routes, PrefixRule{
		Type:        0x00,
		Destination: net.IPNet{IP: ip.Mask(m), Mask: m},
		NextHop:     libol.EthZero,
	})
	for _, rt := range n.Routes {
		_, dest, err := net.ParseCIDR(rt.Prefix)
		if err != nil {
			continue
		}
		nxt := net.ParseIP(rt.NextHop)
		w.routes = append(w.routes, PrefixRule{
			Type:        0x01,
			Destination: *dest,
			NextHop:     nxt,
		})
	}
	return nil
}

func (w *Worker) FreeIpAddr() {
	if w.network == nil {
		return
	}
	if w.listener.DelRoutes != nil {
		_ = w.listener.DelRoutes(w.network.Routes)
	}
	if w.listener.DelAddr != nil {
		prefix := libol.Netmask2Len(w.network.Netmask)
		ipStr := fmt.Sprintf("%s/%d", w.network.IfAddr, prefix)
		_ = w.listener.DelAddr(ipStr)
	}
	w.network = nil
	w.routes = make([]PrefixRule, 0, 32)
}

func (w *Worker) OnClose(s *SocketWorker) error {
	w.out.Info("Worker.OnClose")
	w.FreeIpAddr()
	return nil
}

func (w *Worker) OnSuccess(s *SocketWorker) error {
	w.out.Info("Worker.OnSuccess")
	if w.listener.AddAddr != nil {
		_ = w.listener.AddAddr(w.ifAddr)
	}
	return nil
}

func (w *Worker) UUID() string {
	if w.uuid == "" {
		w.uuid = libol.GenToken(32)
	}
	return w.uuid
}

func (w *Worker) SetUUID(v string) {
	w.uuid = v
}
