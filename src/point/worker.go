package point

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/point/http"
	"net"
	"strings"
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

type socketEvent struct {
	Type   string
	Reason string
	Time   int64
}

func (e socketEvent) String() string {
	return e.Type + " " + e.Reason
}

func NewEvent(typ, reason string) socketEvent {
	return socketEvent{
		Time:   time.Now().Unix(),
		Reason: reason,
		Type:   typ,
	}
}

type jobTimer struct {
	Time int64
	Call func() error
}

type recordTime struct {
	last      int64 // record time last frame received or connected.
	connected int64 // record last connected time.
	reconnect int64 // record time when triggered reconnected.
	sleeps    int   // record times to control connecting delay.
	closed    int64
	live      int64 // record received pong frame time.
}
type SocketWorker struct {
	// private
	listener   SocketWorkerListener
	client     libol.SocketClient
	lock       libol.Locker
	user       *models.User
	network    *models.Network
	routes     map[string]*models.Route
	keepalive  KeepAlive
	done       chan bool
	ticker     *time.Ticker
	pointCfg   *config.Point
	eventQueue chan socketEvent
	writeQueue chan *libol.FrameMessage
	jobber     []jobTimer
	record     recordTime
}

func NewSocketWorker(client libol.SocketClient, c *config.Point) (t *SocketWorker) {
	t = &SocketWorker{
		client:  client,
		user:    models.NewUser(c.Username, c.Password),
		network: models.NewNetwork(c.Network, c.Interface.Address),
		routes:  make(map[string]*models.Route, 64),
		record: recordTime{
			last:      time.Now().Unix(),
			reconnect: time.Now().Unix(),
		},
		done:   make(chan bool, 2),
		ticker: time.NewTicker(2 * time.Second),
		keepalive: KeepAlive{
			Interval: 10,
			LastTime: time.Now().Unix(),
		},
		pointCfg:   c,
		eventQueue: make(chan socketEvent, 32),
		writeQueue: make(chan *libol.FrameMessage, 1024),
		jobber:     make([]jobTimer, 0, 32),
	}
	t.user.Alias = c.Alias
	t.user.Network = c.Network

	return
}

func (t *SocketWorker) sleepNow() int64 {
	return int64(t.record.sleeps * 5)
}

func (t *SocketWorker) sleepIdle() int64 {
	if t.record.sleeps < 20 {
		t.record.sleeps++
	}
	return t.sleepNow()
}

func (t *SocketWorker) Initialize() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.client == nil {
		return
	}
	libol.Info("SocketWorker.Initialize")
	t.client.SetMaxSize(t.pointCfg.Interface.IfMtu)
	t.client.SetListener(libol.ClientListener{
		OnConnected: func(client libol.SocketClient) error {
			t.record.connected = time.Now().Unix()
			t.eventQueue <- NewEvent(EventConed, "from socket")
			return nil
		},
		OnClose: func(client libol.SocketClient) error {
			t.record.closed = time.Now().Unix()
			t.eventQueue <- NewEvent(EventClosed, "from socket")
			return nil
		},
	})
}

func (t *SocketWorker) Start() {
	t.lock.Lock()
	defer t.lock.Unlock()
	_ = t.connect()
	libol.Go(t.Loop)
}

func (t *SocketWorker) leave() {
	if t.client == nil {
		return
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
		libol.Error("SocketWorker.leave: %s", err)
		return
	}
	libol.Cmd("SocketWorker.leave: left: %s", body)
	if err := t.client.WriteReq("left", string(body)); err != nil {
		libol.Error("Switch.leave: %s", err)
		return
	}
}

func (t *SocketWorker) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()

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
	libol.Warn("SocketWorker.connect %s:%d", t.client, libol.ClInit)
	t.client.Close()
	s := t.client.Status()
	if s != libol.ClInit {
		libol.Warn("SocketWorker.connect %s %d->%d", t.client, s, libol.ClInit)
		t.client.SetStatus(libol.ClInit)
	}
	if err := t.client.Connect(); err != nil {
		libol.Error("SocketWorker.connect %s %s", t.client, err)
		return err
	}
	return nil
}

func (t *SocketWorker) reconnect() {
	if t.isStopped() {
		return
	}
	t.record.reconnect = time.Now().Unix()
	t.jobber = append(t.jobber, jobTimer{
		Time: time.Now().Unix() + t.sleepIdle(),
		Call: func() error {
			libol.Debug("SocketWorker.reconnect: on jobber")
			if t.record.connected < t.record.reconnect { // already connected after.
				return t.connect()
			} else {
				libol.Cmd("SocketWorker.reconnect: dissed by already")
			}
			return nil
		},
	})
}

// toLogin request
func (t *SocketWorker) toLogin(client libol.SocketClient) error {
	body, err := json.Marshal(t.user)
	if err != nil {
		libol.Error("SocketWorker.toLogin: %s", err)
		return err
	}
	libol.Cmd("SocketWorker.toLogin: %s", body)
	if err := client.WriteReq("login", string(body)); err != nil {
		libol.Error("SocketWorker.toLogin: %s", err)
		return err
	}
	return nil
}

// network request
func (t *SocketWorker) toNetwork(client libol.SocketClient) error {
	if !t.pointCfg.RequestAddr && t.network.IfAddr == "" {
		libol.Info("SocketWorker.toNetwork: notNeed")
		return nil
	}
	body, err := json.Marshal(t.network)
	if err != nil {
		libol.Error("SocketWorker.toNetwork: %s", err)
		return err
	}
	libol.Cmd("SocketWorker.toNetwork: %s", body)
	if err := client.WriteReq("ipaddr", string(body)); err != nil {
		libol.Error("SocketWorker.toNetwork: %s", err)
		return err
	}
	return nil
}

func (t *SocketWorker) onLogin(resp string) error {
	if strings.HasPrefix(resp, "okay") {
		t.client.SetStatus(libol.ClAuth)
		if t.listener.OnSuccess != nil {
			_ = t.listener.OnSuccess(t)
		}
		t.record.sleeps = 0
		_ = t.toNetwork(t.client)
		t.eventQueue <- NewEvent(EventSuccess, "already success")
		libol.Info("SocketWorker.onInstruct.toLogin: success")
	} else {
		t.client.SetStatus(libol.ClUnAuth)
		libol.Error("SocketWorker.onInstruct.toLogin: %s", resp)
	}
	return nil
}

func (t *SocketWorker) onIpAddr(resp string) error {
	if !t.pointCfg.RequestAddr {
		libol.Info("SocketWorker.onIpAddr: not allowed")
		return nil
	}
	n := &models.Network{}
	if err := json.Unmarshal([]byte(resp), n); err != nil {
		return libol.NewErr("SocketWorker.onInstruct: Invalid json data.")
	}
	t.network = n
	if t.listener.OnIpAddr != nil {
		_ = t.listener.OnIpAddr(t, n)
	}
	return nil
}

func (t *SocketWorker) onLeft(resp string) error {
	client := t.client
	libol.Info("SocketWorker.onLeft: %s %s", client.String(), resp)
	t.close()
	return nil
}

func (t *SocketWorker) onSignIn(resp string) error {
	client := t.client
	libol.Info("SocketWorker.onSignIn: %s %s", client.String(), resp)
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
		libol.Cmd("SocketWorker.onInstruct %s %s", action, resp)
	}
	switch action {
	case "logi:":
		return t.onLogin(resp)
	case "ipad:":
		return t.onIpAddr(resp)
	case "pong:":
		t.record.live = time.Now().Unix()
	case "sign=":
		return t.onSignIn(resp)
	case "left=":
		return t.onLeft(resp)
	default:
		libol.Warn("SocketWorker.onInstruct: %s %s", action, resp)
	}
	return nil
}

func (t *SocketWorker) doTicker() error {
	if t.keepalive.Should() {
		if t.client == nil {
			return nil
		}
		t.keepalive.Update()
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
			libol.Error("SocketWorker.doTicker: %s", err)
			return err
		}
		libol.Cmd("SocketWorker.doTicker: ping= %s", body)
		if err := t.client.WriteReq("ping", string(body)); err != nil {
			libol.Error("SocketWorker.doTicker: %s", err)
			return err
		}
	}

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
	libol.Debug("SocketWorker.doTicker %d", len(t.jobber))
	return nil
}

func (t *SocketWorker) dispatch(ev socketEvent) {
	libol.Info("SocketWorker.dispatch %v", ev)
	switch ev.Type {
	case EventConed:
		if t.client != nil {
			libol.Go(t.Read)
			_ = t.toLogin(t.client)
		}
	case EventSuccess:
	case EventRecon, EventSignIn, EventLogin:
		t.reconnect()
	}
}

func (t *SocketWorker) Loop() {
	defer libol.Info("SocketWorker.Loop exit")
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
			libol.Log("SocketWorker.Ticker: at %s", c)
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
	libol.Info("SocketWorker.Read: %s", t.client)
	for {
		t.lock.Lock()
		if t.isStopped() || !t.client.IsOk() {
			libol.Error("SocketWorker.Read: %v", t.client)
			t.lock.Unlock()
			break
		}
		t.lock.Unlock()
		data, err := t.client.ReadMsg()
		t.lock.Lock()
		if err != nil {
			libol.Error("SocketWorker.Read: %s", err)
			t.lock.Unlock()
			break
		}
		if libol.HasLog(libol.DEBUG) {
			libol.Debug("SocketWorker.Read: %x", data)
		}
		t.record.last = time.Now().Unix()
		if data.Size() > 0 {
			data.Decode()
			if data.IsControl() {
				_ = t.onInstruct(data)
			} else if t.listener.ReadAt != nil {
				_ = t.listener.ReadAt(data)
			}
		}
		t.lock.Unlock()
	}
	if !t.isStopped() {
		t.eventQueue <- NewEvent(EventRecon, "from read")
	}
	libol.Info("SocketWorker.Read: exit")
}

func (t *SocketWorker) deadCheck() {
	dt := time.Now().Unix() - t.record.last
	if dt > int64(t.pointCfg.Timeout) {
		libol.Warn("SocketWorker.deadCheck: %s idle %ds", t.client, dt)
		t.eventQueue <- NewEvent(EventRecon, "from dead check")
		t.record.last = time.Now().Unix()
	}
}

func (t *SocketWorker) DoWrite(frame *libol.FrameMessage) error {
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("SocketWorker.DoWrite: %x", frame)
	}
	t.lock.Lock()
	t.deadCheck()
	if t.client == nil {
		t.lock.Unlock()
		return libol.NewErr("client is nil")
	}
	if t.client.Status() != libol.ClAuth {
		libol.Debug("SocketWorker.Loop: dropping by unAuth")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()
	if err := t.client.WriteMsg(frame); err != nil {
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

func (t *SocketWorker) SetAddr(addr string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.client.SetAddr(addr)
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
	lock       libol.Locker
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
}

func NewTapWorker(devCfg network.TapConfig, c *config.Point) (a *TapWorker) {
	a = &TapWorker{
		device:     nil,
		deviceCfg:  devCfg,
		pointCfg:   c,
		openAgain:  false,
		done:       make(chan bool, 2),
		writeQueue: make(chan *libol.FrameMessage, 1024),
	}
	return
}

func (a *TapWorker) Initialize() {
	a.lock.Lock()
	defer a.lock.Unlock()

	libol.Info("TapWorker.Initialize")
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
		libol.Info("TapWorker.Initialize: src %x", a.ether.HwAddr)
	}
}

func (a *TapWorker) setEther(addr string) {
	a.neighbor.Clear()
	// format ip address.
	addr = libol.IpAddrFormat(addr)
	ifAddr := strings.SplitN(addr, "/", 2)[0]
	a.ether.IpAddr = net.ParseIP(ifAddr).To4()
	if a.ether.IpAddr == nil {
		libol.Warn("TapWorker.setEther: srcIp is nil")
		a.ether.IpAddr = []byte{0x00, 0x00, 0x00, 0x00}
	} else {
		libol.Info("TapWorker.setEther: srcIp % x", a.ether.IpAddr)
	}
	// changed address need open device again.
	if a.ifAddr != "" && a.ifAddr != addr {
		libol.Warn("TapWorker.setEther changed %s->%s", a.ifAddr, addr)
		a.openAgain = true
	}
	a.ifAddr = addr
}

func (a *TapWorker) open() {
	if a.device != nil {
		_ = a.device.Close()
		if !a.openAgain {
			time.Sleep(5 * time.Second) // sleep 5s and release cpu.
		}
	}
	device, err := network.NewKernelTap(a.pointCfg.Network, a.deviceCfg)
	if err != nil {
		libol.Error("TapWorker.open: %s", err)
		return
	}
	libol.Info("TapWorker.open: >>>> %s <<<<", device.Name())
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
	libol.Debug("TapWorker.onMiss: %v.", dest)
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
	libol.Debug("TapWorker.onMiss: %x.", frame.Frame()[:64])
	if a.listener.ReadAt != nil {
		_ = a.listener.ReadAt(frame)
	}
}

func (a *TapWorker) onFrame(frame *libol.FrameMessage, data []byte) int {
	size := len(data)
	if a.device.IsTun() {
		iph, err := libol.NewIpv4FromFrame(data)
		if err != nil {
			libol.Warn("TapWorker.onFrame: %s", err)
			return 0
		}
		dest := iph.Destination
		if a.listener.FindNext != nil {
			dest = a.listener.FindNext(dest)
		}
		neb := a.neighbor.GetByBytes(dest)
		if neb == nil {
			a.onMiss(dest)
			libol.Debug("TapWorker.onFrame: onMiss neighbor %v", dest)
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
	defer libol.Catch("TapWorker.Read")

	libol.Info("TapWorker.Read")
	for {
		a.lock.Lock()
		if a.device == nil {
			a.close()
			a.lock.Unlock()
			break
		}
		a.lock.Unlock()

		frame := libol.NewFrameMessage()
		data := frame.Frame()
		if a.device.IsTun() {
			data = data[libol.EtherLen:]
		}
		n, err := a.device.Read(data)
		a.lock.Lock()
		if err != nil || a.openAgain {
			if err != nil {
				libol.Warn("TapWorker.Read: %s", err)
			}
			a.open()
			a.openAgain = false // clear openAgain flags
			a.lock.Unlock()
			continue
		}
		if libol.HasLog(libol.DEBUG) {
			libol.Debug("TapWorker.Read: %x", data[:n])
		}
		if size := a.onFrame(frame, data[:n]); size == 0 {
			a.lock.Unlock()
			continue
		}
		if a.listener.ReadAt != nil {
			_ = a.listener.ReadAt(frame)
		}
		a.lock.Unlock()
	}
	libol.Info("TapWorker.Read: exit")
}

func (a *TapWorker) Loop() {
	defer libol.Info("TapWorker.Loop exit")
	for {
		select {
		case <-a.done:
			return
		case d := <-a.writeQueue:
			_ = a.DoWrite(d)
		}
	}
}

func (a *TapWorker) DoWrite(frame *libol.FrameMessage) error {
	data := frame.Frame()
	if libol.HasLog(libol.DEBUG) {
		libol.Debug("TapWorker.DoWrite: %x", data)
	}
	a.lock.Lock()
	if a.device == nil {
		a.lock.Unlock()
		return libol.NewErr("device is nil")
	}
	if a.device.IsTun() {
		// proxy arp request.
		if a.toArp(data) {
			libol.Debug("TapWorker.Loop: Arp proxy.")
			a.lock.Unlock()
			return nil
		}
		eth, err := libol.NewEtherFromFrame(data)
		if err != nil {
			libol.Error("TapWorker.Loop: %s", err)
			a.lock.Unlock()
			return nil
		}
		if eth.IsIP4() {
			data = data[14:]
		} else {
			libol.Debug("TapWorker.Loop: 0x%04x not IPv4", eth.Type)
			a.lock.Unlock()
			return nil
		}
	}
	a.lock.Unlock()
	if _, err := a.device.Write(data); err != nil {
		libol.Error("TapWorker.Loop: %s", err)
		a.lock.Lock()
		a.close()
		a.lock.Unlock()
		return err
	}
	return nil
}

// learn source from arp
func (a *TapWorker) toArp(data []byte) bool {
	libol.Debug("TapWorker.toArp")
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("TapWorker.toArp: %s", err)
		return false
	}
	if !eth.IsArp() {
		return false
	}
	arp, err := libol.NewArpFromFrame(data[eth.Len:])
	if err != nil {
		libol.Error("TapWorker.toArp: %s.", err)
		return false
	}
	if arp.IsIP4() {
		if !bytes.Equal(eth.Src, arp.SHwAddr) {
			libol.Error("TapWorker.toArp: eth.dst not arp.shw %x.", arp.SIpAddr)
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
				libol.Cmd1("TapWorker.toArp: reply %v on %x.", rep.SIpAddr, rep.SHwAddr)
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
				libol.Cmd1("TapWorker.toArp: recv %v on %x.", arp.SIpAddr, arp.SHwAddr)
			}
		default:
			libol.Warn("TapWorker.toArp: not op %x.", arp.OpCode)
		}
	}
	return true
}

func (a *TapWorker) close() {
	libol.Info("TapWorker.close")
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
	libol.Go(a.Read)
	libol.Go(a.Loop)
	libol.Go(a.neighbor.Start)
}

func (a *TapWorker) Stop() {
	a.lock.Lock()
	defer a.lock.Unlock()
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
		kcpCfg := &libol.KcpConfig{
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewKcpClient(c.Connection, kcpCfg)
	case "tcp":
		tcpCfg := &libol.TcpConfig{
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewTcpClient(c.Connection, tcpCfg)
	case "udp":
		udpCfg := &libol.UdpConfig{
			Block:   config.GetBlock(c.Crypt),
			Timeout: time.Duration(c.Timeout) * time.Second,
		}
		return libol.NewUdpClient(c.Connection, udpCfg)
	default:
		tcpCfg := &libol.TcpConfig{
			Tls:   &tls.Config{InsecureSkipVerify: true},
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewTcpClient(c.Connection, tcpCfg)
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
	tcpWorker *SocketWorker
	tapWorker *TapWorker
	config    *config.Point
	uuid      string
	network   *models.Network
	routes    []PrefixRule
}

func NewWorker(config *config.Point) (p *Worker) {
	return &Worker{
		ifAddr: config.Interface.Address,
		config: config,
		routes: make([]PrefixRule, 0, 32),
	}
}

func (p *Worker) Initialize() {
	if p.config == nil {
		return
	}
	libol.Info("Worker.Initialize")
	client := GetSocketClient(p.config)
	p.tcpWorker = NewSocketWorker(client, p.config)

	tapCfg := GetTapCfg(p.config)
	// register listener
	p.tapWorker = NewTapWorker(tapCfg, p.config)

	p.tcpWorker.SetUUID(p.UUID())
	p.tcpWorker.listener = SocketWorkerListener{
		OnClose:   p.OnClose,
		OnSuccess: p.OnSuccess,
		OnIpAddr:  p.OnIpAddr,
		ReadAt: func(frame *libol.FrameMessage) error {
			p.tapWorker.writeQueue <- frame
			return nil
		},
	}
	p.tcpWorker.Initialize()

	p.tapWorker.listener = TapWorkerListener{
		OnOpen: func(w *TapWorker) error {
			if p.listener.OnTap != nil {
				if err := p.listener.OnTap(w); err != nil {
					return err
				}
			}
			if p.network != nil {
				_ = p.OnIpAddr(p.tcpWorker, p.network)
			}
			return nil
		},
		ReadAt: func(frame *libol.FrameMessage) error {
			p.tcpWorker.writeQueue <- frame
			return nil
		},
		FindNext: p.FindNext,
	}
	p.tapWorker.Initialize()

	if p.config.Http != nil {
		p.http = http.NewHttp(p)
	}
}

func (p *Worker) Start() {
	libol.Debug("Worker.Start linux.")
	p.tapWorker.Start()
	p.tcpWorker.Start()
	if p.http != nil {
		libol.Go(p.http.Start)
	}
}

func (p *Worker) Stop() {
	if p.tapWorker == nil || p.tcpWorker == nil {
		return
	}
	if p.http != nil {
		p.http.Shutdown()
	}
	p.FreeIpAddr()
	p.tcpWorker.Stop()
	p.tapWorker.Stop()
	p.tcpWorker = nil
	p.tapWorker = nil
}

func (p *Worker) Client() libol.SocketClient {
	if p.tcpWorker != nil {
		return p.tcpWorker.client
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
		return client.Addr()
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
	if p.tcpWorker != nil {
		return p.tcpWorker
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
		if libol.HasLog(libol.DEBUG) {
			libol.Debug("Worker.FindNext %v to %v", dest, rt.NextHop)
		}
		return rt.NextHop.To4()
	}
	return dest
}

func (p *Worker) OnIpAddr(w *SocketWorker, n *models.Network) error {
	libol.Info("Worker.OnIpAddr: %s/%s, %s", n.IfAddr, n.Netmask, n.Routes)

	if p.network != nil { // remove older firstly
		p.FreeIpAddr()
	}
	prefix := libol.Netmask2Len(n.Netmask)
	ipStr := fmt.Sprintf("%s/%d", n.IfAddr, prefix)
	p.tapWorker.setEther(ipStr)
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
	libol.Info("Worker.OnClose")
	p.FreeIpAddr()
	return nil
}

func (p *Worker) OnSuccess(w *SocketWorker) error {
	libol.Info("Worker.OnSuccess")
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
	return p.config
}
