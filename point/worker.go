package point

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/network"
	"github.com/danieldin95/openlan-go/point/http"
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
	ReadAt    func(p []byte) error
}

type SocketWorker struct {
	// public
	Listener SocketWorkerListener
	Client   libol.SocketClient
	// private
	lock        sync.RWMutex
	user        *models.User
	network     *models.Network
	routes      map[string]*models.Route
	initialized bool
	lastTime    int64 // 15s timeout.
	sleepTimes  int
	keepalive   KeepAlive
	done        chan bool
	ticker      *time.Ticker
	pointCfg    *config.Point
}

func NewSocketWorker(client libol.SocketClient, c *config.Point) (t *SocketWorker) {
	t = &SocketWorker{
		Client:      client,
		user:        models.NewUser(c.Username, c.Password),
		network:     models.NewNetwork(c.Network, c.Intf.Address),
		routes:      make(map[string]*models.Route, 64),
		initialized: false,
		lastTime:    time.Now().Unix(),
		done:        make(chan bool),
		ticker:      time.NewTicker(5 * time.Second),
		keepalive: KeepAlive{
			Interval: 10,
			LastTime: time.Now().Unix(),
		},
		pointCfg: c,
	}
	t.user.Alias = c.Alias
	t.user.Network = c.Network

	return
}

func (t *SocketWorker) GetSleepIdle() time.Duration {
	if t.sleepTimes < 20 {
		t.sleepTimes++
	}
	return time.Duration(t.sleepTimes) * time.Second * 3
}

func (t *SocketWorker) Initialize() {
	if t.Client == nil {
		return
	}
	libol.Info("SocketWorker.Initialize")
	t.initialized = true
	t.Client.SetMaxSize(t.pointCfg.Intf.Mtu)
	t.Client.SetListener(libol.ClientListener{
		OnConnected: func(client libol.SocketClient) error {
			return t.Login(client)
		},
		OnClose: func(client libol.SocketClient) error {
			if t.Listener.OnClose != nil {
				_ = t.Listener.OnClose(t)
			}
			return nil
		},
	})
}

func (t *SocketWorker) Start() {
	if !t.initialized {
		t.Initialize()
	}
	_ = t.Connect()
	go t.Read()
	go t.Loop()
}

func (t *SocketWorker) Leave() {
	if t.Client == nil {
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
		Address:    t.Client.LocalAddr(),
		Connection: t.Client.RemoteAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		libol.Error("SocketWorker.Leave: %s", err)
		return
	}
	libol.Cmd("SocketWorker.Leave: left: %s", body)
	if err := t.Client.WriteReq("left", string(body)); err != nil {
		libol.Error("Switch.Leave: %s", err)
		return
	}
}

func (t *SocketWorker) Stop() {
	t.Leave()
	t.Client.Terminal()

	t.lock.Lock()
	defer t.lock.Unlock()
	t.Client = nil

	t.ticker.Stop()
	t.done <- true
}

func (t *SocketWorker) Close() {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.Client != nil {
		t.Client.Close()
	}
}

func (t *SocketWorker) Connect() error {
	s := t.Client.Status()
	if s != libol.ClInit {
		libol.Warn("SocketWorker.Connect %s %d->%d", t.Client, s, libol.ClInit)
		t.Client.SetStatus(libol.ClInit)
	}
	if err := t.Client.Connect(); err != nil {
		libol.Error("SocketWorker.Connect %s %s", t.Client, err)
		return err
	}
	return nil
}

// login request
func (t *SocketWorker) Login(client libol.SocketClient) error {
	body, err := json.Marshal(t.user)
	if err != nil {
		libol.Error("SocketWorker.Login: %s", err)
		return err
	}
	libol.Cmd("SocketWorker.Login: %s", body)
	if err := client.WriteReq("login", string(body)); err != nil {
		libol.Error("SocketWorker.Login: %s", err)
		return err
	}
	return nil
}

// network request
func (t *SocketWorker) Network(client libol.SocketClient) error {
	body, err := json.Marshal(t.network)
	if err != nil {
		libol.Error("SocketWorker.Network: %s", err)
		return err
	}
	libol.Cmd("SocketWorker.Network: %s", body)
	if err := client.WriteReq("ipaddr", string(body)); err != nil {
		libol.Error("SocketWorker.Network: %s", err)
		return err
	}
	return nil
}

// handle instruct from virtual switch
func (t *SocketWorker) onInstruct(data []byte) error {
	m := libol.NewFrameMessage(data)
	if !m.IsControl() {
		return nil
	}
	action, resp := m.CmdAndParams()
	libol.Cmd("SocketWorker.onInstruct %s %s", action, resp)
	switch action {
	case "logi:":
		if strings.HasPrefix(resp, "okay") {
			t.Client.SetStatus(libol.ClAuth)
			if t.Listener.OnSuccess != nil {
				_ = t.Listener.OnSuccess(t)
			}
			t.sleepTimes = 0
			if t.pointCfg.RequestAddr {
				_ = t.Network(t.Client)
			}
			libol.Info("SocketWorker.onInstruct.login: success")
		} else {
			t.Client.SetStatus(libol.ClUnAuth)
			libol.Error("SocketWorker.onInstruct.login: %s", resp)
		}
	case "ipad:":
		n := models.Network{}
		if err := json.Unmarshal([]byte(resp), &n); err != nil {
			return libol.NewErr("SocketWorker.onInstruct: Invalid json data.")
		}
		if t.Listener.OnIpAddr != nil {
			_ = t.Listener.OnIpAddr(t, &n)
		}
	case "sign=":
		_ = t.Login(t.Client)
	}
	return nil
}

func (t *SocketWorker) Ticker() error {
	if t.keepalive.Should() {
		if t.Client == nil {
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
			Address:    t.Client.LocalAddr(),
			Connection: t.Client.RemoteAddr(),
		}
		body, err := json.Marshal(data)
		if err != nil {
			libol.Error("SocketWorker.Ticker: %s", err)
			return err
		}
		libol.Cmd("SocketWorker.Ticker: ping: %s", body)
		if err := t.Client.WriteReq("ping", string(body)); err != nil {
			libol.Error("SocketWorker.Ticker: %s", err)
			return err
		}
	}
	return nil
}

func (t *SocketWorker) Loop() {
	defer libol.Info("SocketWorker.Loop exit")
	for {
		select {
		case <-t.done:
			return
		case c := <-t.ticker.C:
			libol.Log("Neighbors.Expire: tick at %s", c)
			_ = t.Ticker()
		}
	}
}

func (t *SocketWorker) Read() {
	libol.Info("SocketWorker.Read: %s", t.Client.State())
	defer libol.Catch("SocketWorker.Read")

	data := make([]byte, libol.MAXBUF)
	for {
		if t.Client == nil || t.Client.Have(libol.ClTerminal) {
			break
		}
		if !t.Client.IsOk() {
			time.Sleep(t.GetSleepIdle()) // sleep 30s and release cpu.
			_ = t.Connect()
			continue
		}
		n, err := t.Client.ReadMsg(data)
		if err != nil {
			libol.Error("SocketWorker.Read: %s", err)
			t.Close()
			continue
		}
		t.lastTime = time.Now().Unix()
		libol.Log("SocketWorker.Read: %x", data[:n])
		if n > 0 {
			frame := data[:n]
			if libol.IsControl(frame) {
				_ = t.onInstruct(frame)
			} else if t.Listener.ReadAt != nil {
				_ = t.Listener.ReadAt(frame)
			}
		}
	}
	t.Close()
	libol.Info("SocketWorker.Read: exit")
}

func (t *SocketWorker) DeadCheck() {
	dt := time.Now().Unix() - t.lastTime
	if dt > int64(t.pointCfg.Timeout) {
		libol.Warn("SocketWorker.DeadCheck: %s idle %ds", t.Client.String(), dt)
		t.Close()
		_ = t.Connect()
		//
		t.lastTime = time.Now().Unix()
	}
}

func (t *SocketWorker) DoWrite(data []byte) error {
	libol.Log("SocketWorker.DoWrite: %x", data)

	t.DeadCheck()
	if t.Client == nil {
		return libol.NewErr("Client is nil")
	}
	if t.Client.Status() != libol.ClAuth {
		libol.Debug("SocketWorker.Loop: dropping by unAuth")
		return nil
	}
	if err := t.Client.WriteMsg(data); err != nil {
		t.Close()
		return err
	}
	return nil
}

func (t *SocketWorker) Auth() (string, string) {
	return t.user.Name, t.user.Password
}

func (t *SocketWorker) SetAuth(auth string) {
	values := strings.Split(auth, ":")
	t.user.Name = values[0]
	if len(values) > 1 {
		t.user.Password = values[1]
	}
}

func (t *SocketWorker) SetAddr(addr string) {
	t.Client.SetAddr(addr)
}

func (t *SocketWorker) SetUUID(v string) {
	t.user.UUID = v
}

type TapWorkerListener struct {
	OnOpen   func(w *TapWorker) error
	OnClose  func(w *TapWorker)
	FindDest func(dest []byte) []byte
	ReadAt   func([]byte) error
}

type TunEther struct {
	HwAddr []byte
	IpAddr []byte
}

type TapWorker struct {
	// public
	Device    network.Taper
	Listener  TapWorkerListener
	Ether     TunEther
	Neighbors Neighbors
	OpenAgain *libol.SafeVar
	// private
	lock        sync.RWMutex
	devCfg      network.TapConfig
	pointCfg    *config.Point
	initialized bool
	ifAddr      string
}

func NewTapWorker(devCfg network.TapConfig, c *config.Point) (a *TapWorker) {
	a = &TapWorker{
		Device:      nil,
		devCfg:      devCfg,
		pointCfg:    c,
		initialized: false,
		OpenAgain:   libol.NewSafeVar(),
	}
	a.OpenAgain.Set(false)

	return
}

func (a *TapWorker) Initialize() {
	libol.Info("TapWorker.Initialize")
	a.initialized = true
	a.Neighbors = Neighbors{
		neighbors: make(map[uint32]*Neighbor, 1024),
		done:      make(chan bool),
		ticker:    time.NewTicker(5 * time.Second),
		timeout:   5 * 60,
	}
	a.Open()
	a.DoTun()
}

func (a *TapWorker) SetEther(addr string) {
	ifAddr := strings.SplitN(addr, "/", 2)[0]
	if ifAddr != "" {
		a.Ether.IpAddr = net.ParseIP(ifAddr).To4()
		if a.Ether.IpAddr == nil {
			libol.Warn("TapWorker.SetEther: srcIp is nil")
			a.Ether.IpAddr = []byte{0x00, 0x00, 0x00, 0x00}
		} else {
			libol.Info("TapWorker.SetEther: srcIp % x", a.Ether.IpAddr)
		}
	}
	// changed address need open device again.
	if a.ifAddr != "" && a.ifAddr != addr {
		libol.Warn("TapWorker.SetEther changed %s->%s", a.ifAddr, addr)
		a.OpenAgain.Set(true)
	}
	a.ifAddr = addr
	a.Neighbors.Clear()
}

func (a *TapWorker) DoTun() {
	if a.Device == nil || !a.Device.IsTun() {
		return
	}
	a.SetEther(a.pointCfg.Intf.Address)
	a.Ether.HwAddr = libol.GenEthAddr(6)
	libol.Info("TapWorker.DoTun: src %x", a.Ether.HwAddr)
}

func (a *TapWorker) Open() {
	if a.Device != nil {
		_ = a.Device.Close()
		if !a.OpenAgain.Get().(bool) {
			time.Sleep(5 * time.Second) // sleep 5s and release cpu.
		}
	}
	dev, err := network.NewKernelTap(a.pointCfg.Network, a.devCfg)
	if err != nil {
		libol.Error("TapWorker.Open: %s", err)
		return
	}
	libol.Info("TapWorker.Open: >>>> %s <<<<", dev.Name())
	a.Device = dev
	if a.Listener.OnOpen != nil {
		_ = a.Listener.OnOpen(a)
	}
}

func (a *TapWorker) NewEth(t uint16, dst []byte) *libol.Ether {
	eth := libol.NewEther(t)
	eth.Dst = dst
	eth.Src = a.Ether.HwAddr
	return eth
}

// process if ethernet destination is missed
func (a *TapWorker) onMiss(dest []byte) {
	libol.Debug("TapWorker.onMiss: %x.", dest)
	eth := a.NewEth(libol.EthArp, libol.BROADED)
	reply := libol.NewArp()
	reply.OpCode = libol.ArpRequest
	reply.SIpAddr = a.Ether.IpAddr
	reply.TIpAddr = dest
	reply.SHwAddr = a.Ether.HwAddr
	reply.THwAddr = libol.ZEROED

	buffer := make([]byte, 0, a.pointCfg.Intf.Mtu)
	buffer = append(buffer, eth.Encode()...)
	buffer = append(buffer, reply.Encode()...)

	libol.Debug("TapWorker.onMiss: %x.", buffer)
	if a.Listener.ReadAt != nil {
		_ = a.Listener.ReadAt(buffer)
	}
}

func (a *TapWorker) Read() {
	defer libol.Catch("TapWorker.Read")

	libol.Info("TapWorker.Read")
	data := make([]byte, libol.MAXBUF)
	for {
		if a.Device == nil {
			break
		}

		n, err := a.Device.Read(data)
		if err != nil || a.OpenAgain.Get().(bool) {
			if err != nil {
				libol.Warn("TapWorker.Read: %s", err)
			}
			a.Open()
			// clear OpenAgain flags
			a.OpenAgain.Set(false)
			continue
		}
		libol.Log("TapWorker.Read: %x", data[:n])
		if a.Device.IsTun() {
			iph, err := libol.NewIpv4FromFrame(data)
			if err != nil {
				libol.Error("TapWorker.Read: %s", err)
				continue
			}
			dest := iph.Destination
			if a.Listener.FindDest != nil {
				dest = a.Listener.FindDest(dest)
			}
			neb := a.Neighbors.GetByBytes(dest)
			if neb == nil {
				a.onMiss(dest)
				continue
			}
			eth := a.NewEth(libol.EthIp4, neb.HwAddr)
			buffer := make([]byte, 0, libol.MAXBUF)
			buffer = append(buffer, eth.Encode()...)
			buffer = append(buffer, data[0:n]...)
			n += eth.Len
			if a.Listener.ReadAt != nil {
				_ = a.Listener.ReadAt(buffer[:n])
			}
		} else {
			if a.Listener.ReadAt != nil {
				_ = a.Listener.ReadAt(data[:n])
			}
		}
	}
	a.Close()
	libol.Info("TapWorker.Read: exit")
}

func (a *TapWorker) DoWrite(data []byte) error {
	libol.Log("TapWorker.DoWrite: %x", data)

	if a.Device == nil {
		return libol.NewErr("Device is nil")
	}

	if a.Device.IsTun() {
		//Proxy arp request.
		if a.onArp(data) {
			libol.Debug("TapWorker.Loop: Arp proxy.")
			return nil
		}
		eth, err := libol.NewEtherFromFrame(data)
		if err != nil {
			libol.Error("TapWorker.Loop: %s", err)
			return nil
		}
		if eth.IsIP4() {
			data = data[14:]
		} else {
			libol.Debug("TapWorker.Loop: 0x%04x not IPv4", eth.Type)
			return nil
		}
	}

	if _, err := a.Device.Write(data); err != nil {
		libol.Error("TapWorker.Loop: %s", err)
		a.Close()
		return err
	}
	return nil
}

// learn source from arp
func (a *TapWorker) onArp(data []byte) bool {
	libol.Debug("TapWorker.onArp")
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("TapWorker.onArp: %s", err)
		return false
	}
	if !eth.IsArp() {
		return false
	}

	arp, err := libol.NewArpFromFrame(data[eth.Len:])
	if err != nil {
		libol.Error("TapWorker.onArp: %s.", err)
		return false
	}
	if arp.IsIP4() {
		if !bytes.Equal(eth.Src, arp.SHwAddr) {
			libol.Error("TapWorker.onArp: eth.dst not arp.shw %x.", arp.SIpAddr)
			return true
		}
		if arp.OpCode == libol.ArpRequest && bytes.Equal(arp.TIpAddr, a.Ether.IpAddr) {
			eth := a.NewEth(libol.EthArp, arp.SHwAddr)
			reply := libol.NewArp()
			reply.OpCode = libol.ArpReply
			reply.SIpAddr = a.Ether.IpAddr
			reply.TIpAddr = arp.SIpAddr
			reply.SHwAddr = a.Ether.HwAddr
			reply.THwAddr = arp.SHwAddr
			buffer := make([]byte, 0, a.pointCfg.Intf.Mtu)
			buffer = append(buffer, eth.Encode()...)
			buffer = append(buffer, reply.Encode()...)
			libol.Info("TapWorker.onArp: reply %x.", buffer)
			if a.Listener.ReadAt != nil {
				_ = a.Listener.ReadAt(buffer)
			}
		} else if arp.OpCode == libol.ArpReply && bytes.Equal(arp.THwAddr, a.Ether.HwAddr) {
			a.Neighbors.Add(&Neighbor{
				HwAddr:  arp.SHwAddr,
				IpAddr:  arp.SIpAddr,
				NewTime: time.Now().Unix(),
				Uptime:  time.Now().Unix(),
			})
			libol.Info("TapWorker.onArp: recv %x on %x.", arp.SHwAddr, arp.SIpAddr)
		} else {
			libol.Warn("TapWorker.onArp: not op %x.", arp.OpCode)
		}
	}
	return true
}

func (a *TapWorker) Close() {
	libol.Info("TapWorker.Close")
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.Device != nil {
		if a.Listener.OnClose != nil {
			a.Listener.OnClose(a)
		}
		_ = a.Device.Close()
		a.Device = nil
	}
}

func (a *TapWorker) Start() {
	if !a.initialized {
		a.Initialize()
	}
	go a.Read()
	go a.Neighbors.Start()
}

func (a *TapWorker) Stop() {
	a.Neighbors.Stop()
	a.Close()
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
	var tlsConf *tls.Config
	if c.Protocol == "tls" {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}
	switch c.Protocol {
	case "kcp":
		return libol.NewKcpClient(c.Addr, nil)
	case "tcp":
		return libol.NewTcpClient(c.Addr, nil)
	case "udp":
		return libol.NewUdpClient(c.Addr, nil)
	default:
		return libol.NewTcpClient(c.Addr, tlsConf)
	}
}

func GetTapCfg(c *config.Point) network.TapConfig {
	if c.Intf.Provider == "tun" {
		return network.TapConfig{
			Type:    network.TUN,
			Name:    c.Intf.Name,
			Network: c.Intf.Address,
		}
	} else {
		return network.TapConfig{
			Type:    network.TAP,
			Name:    c.Intf.Name,
			Network: c.Intf.Address,
		}
	}
}

type Worker struct {
	// public
	IfAddr   string
	Listener WorkerListener
	// private
	http        *http.Http
	tcpWorker   *SocketWorker
	tapWorker   *TapWorker
	config      *config.Point
	uuid        string
	network     *models.Network
	initialized bool
	routes      []PrefixRule
}

func NewWorker(config *config.Point) (p *Worker) {
	return &Worker{
		IfAddr:      config.Intf.Address,
		config:      config,
		initialized: false,
		routes:      make([]PrefixRule, 0, 32),
	}
}

func (p *Worker) Initialize() {
	if p.config == nil {
		return
	}
	libol.Info("Worker.Initialize")
	client := GetSocketClient(p.config)
	p.initialized = true
	p.tcpWorker = NewSocketWorker(client, p.config)

	tapCfg := GetTapCfg(p.config)
	// register listener
	p.tapWorker = NewTapWorker(tapCfg, p.config)

	p.tcpWorker.SetUUID(p.UUID())
	p.tcpWorker.Listener = SocketWorkerListener{
		OnClose:   p.OnClose,
		OnSuccess: p.OnSuccess,
		OnIpAddr:  p.OnIpAddr,
		ReadAt:    p.tapWorker.DoWrite,
	}
	p.tcpWorker.Initialize()

	p.tapWorker.Listener = TapWorkerListener{
		OnOpen: func(w *TapWorker) error {
			if p.Listener.OnTap != nil {
				if err := p.Listener.OnTap(w); err != nil {
					return err
				}
			}
			if p.network != nil {
				_ = p.OnIpAddr(p.tcpWorker, p.network)
			}
			return nil
		},
		ReadAt:   p.tcpWorker.DoWrite,
		FindDest: p.FindDest,
	}
	p.tapWorker.Initialize()

	if p.config.Http != nil {
		p.http = http.NewHttp(p)
	}
}

func (p *Worker) Start() {
	libol.Debug("Worker.Start linux.")
	if !p.initialized {
		p.Initialize()
	}
	p.tapWorker.Start()
	p.tcpWorker.Start()

	if p.http != nil {
		_ = p.http.Start()
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
		return p.tcpWorker.Client
	}
	return nil
}

func (p *Worker) Device() network.Taper {
	if p.tapWorker != nil {
		return p.tapWorker.Device
	}
	return nil
}

func (p *Worker) UpTime() int64 {
	client := p.Client()
	if client != nil {
		return client.UpTime()
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

func (p *Worker) FindDest(dest []byte) []byte {
	for _, rt := range p.routes {
		if rt.Destination.Contains(dest) {
			if rt.Type == 0x00 {
				break
			}
			libol.Debug("Worker.FindDest %x to %v", dest, rt.NextHop)
			return rt.NextHop.To4()
		}
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
	p.tapWorker.SetEther(ipStr)
	if p.Listener.AddAddr != nil {
		_ = p.Listener.AddAddr(ipStr)
	}
	if p.Listener.AddRoutes != nil {
		_ = p.Listener.AddRoutes(n.Routes)
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
	if p.Listener.DelRoutes != nil {
		_ = p.Listener.DelRoutes(p.network.Routes)
	}
	if p.Listener.DelAddr != nil {
		prefix := libol.Netmask2Len(p.network.Netmask)
		ipStr := fmt.Sprintf("%s/%d", p.network.IfAddr, prefix)
		_ = p.Listener.DelAddr(ipStr)
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
	if p.Listener.AddAddr != nil {
		_ = p.Listener.AddAddr(p.IfAddr)
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
