package point

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
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

type SessWorkerListener struct {
	OnClose   func(w *SocketWorker) error
	OnSuccess func(w *SocketWorker) error
	OnIpAddr  func(w *SocketWorker, n *models.Network) error
	ReadAt    func(p []byte) error
}

type SocketWorker struct {
	// public
	Listener SessWorkerListener
	Client   libol.SocketClient
	// private
	lock        sync.RWMutex
	maxSize     int
	alias       string
	user        *models.User
	network     *models.Network
	routes      map[string]*models.Route
	allowed     bool
	initialized bool
	idleTime    int64 // 15s timeout.
	sleepTimes  int
}

func NewSessWorker(client libol.SocketClient, c *config.Point) (t *SocketWorker) {
	t = &SocketWorker{
		Client:      client,
		maxSize:     c.If.Mtu,
		user:        models.NewUser(c.Username, c.Password),
		network:     models.NewNetwork(c.Network, c.If.Address),
		routes:      make(map[string]*models.Route, 64),
		allowed:     c.Allowed,
		initialized: false,
		idleTime:    time.Now().Unix(),
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
	t.Client.SetMaxSize(t.maxSize)
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
}

func (t *SocketWorker) Stop() {
	t.Client.Terminal()

	t.lock.Lock()
	defer t.lock.Unlock()
	t.Client = nil
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
	if s != libol.CL_INIT {
		libol.Warn("SocketWorker.Connect status %d->%d", s, libol.CL_INIT)
		t.Client.SetStatus(libol.CL_INIT)
	}
	if err := t.Client.Connect(); err != nil {
		libol.Error("SocketWorker.Connect %s", err)
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
		if resp[:4] == "okay" {
			t.Client.SetStatus(libol.CL_AUEHED)
			if t.Listener.OnSuccess != nil {
				_ = t.Listener.OnSuccess(t)
			}
			t.sleepTimes = 0
			if t.allowed {
				_ = t.Network(t.Client)
			}
			libol.Info("SocketWorker.onInstruct.login: success")
		} else {
			t.Client.SetStatus(libol.CL_UNAUTH)
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

func (t *SocketWorker) Read() {
	libol.Info("SocketWorker.Read: %s", t.Client.State())
	defer libol.Catch("SocketWorker.Read")

	data := make([]byte, libol.MAXBUF)
	for {
		if t.Client == nil || t.Client.Have(libol.CL_TERMINAL) {
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
		t.idleTime = time.Now().Unix()
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

func (t *SocketWorker) Expire() {
	if time.Now().Unix()-t.idleTime > 15 {
		t.Close()
		_ = t.Connect()
	}
}

func (t *SocketWorker) DoWrite(data []byte) error {
	libol.Log("SocketWorker.DoWrite: %x", data)
	//t.Expire()
	if t.Client == nil {
		return libol.NewErr("Client is nil")
	}
	if t.Client.Status() != libol.CL_AUEHED {
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

type Neighbor struct {
	HwAddr  []byte
	IpAddr  []byte
	Uptime  int64
	NewTime int64
}

type Neighbors struct {
	lock      sync.RWMutex
	neighbors map[uint32]*Neighbor
	done      chan bool
	ticker    *time.Ticker
	timeout   int64
}

func (n *Neighbors) Expire() {
	deletes := make([]uint32, 0, 1024)

	n.lock.Lock()
	defer n.lock.Unlock()
	//collect need deleted.
	for index, learn := range n.neighbors {
		now := time.Now().Unix()
		if now-learn.Uptime > n.timeout {
			deletes = append(deletes, index)
		}
	}
	libol.Debug("Neighbors.Expire delete %d", len(deletes))
	//execute delete.
	for _, d := range deletes {
		if l, ok := n.neighbors[d]; ok {
			delete(n.neighbors, d)
			libol.Info("Neighbors.Expire: delete %x", l.HwAddr)
		}
	}
}

func (n *Neighbors) Start() {
	for {
		select {
		case <-n.done:
			return
		case t := <-n.ticker.C:
			libol.Log("Neighbors.Expire: tick at %s", t)
			n.Expire()
		}
	}
}

func (n *Neighbors) Stop() {
	n.ticker.Stop()
	n.done <- true
}

func (n *Neighbors) Add(h *Neighbor) {
	if h == nil {
		return
	}
	n.lock.Lock()
	defer n.lock.Unlock()
	k := binary.BigEndian.Uint32(h.IpAddr)
	if l, ok := n.neighbors[k]; ok {
		l.Uptime = h.Uptime
		copy(l.HwAddr[:6], h.HwAddr[:6])
	} else {
		l := &Neighbor{
			Uptime:  h.Uptime,
			NewTime: h.NewTime,
			HwAddr:  make([]byte, 6),
			IpAddr:  make([]byte, 4),
		}
		copy(l.IpAddr[:4], h.IpAddr[:4])
		copy(l.HwAddr[:6], h.HwAddr[:6])
		n.neighbors[k] = l
	}
}

func (n *Neighbors) Get(d uint32) *Neighbor {
	n.lock.RLock()
	defer n.lock.RUnlock()
	if l, ok := n.neighbors[d]; ok {
		return l
	}
	return nil
}

func (n *Neighbors) Clear() {
	libol.Info("Neighbor.Clear")
	n.lock.Lock()
	defer n.lock.Unlock()

	deletes := make([]uint32, 0, 1024)
	for index := range n.neighbors {
		deletes = append(deletes, index)
	}
	//execute delete.
	for _, d := range deletes {
		if _, ok := n.neighbors[d]; ok {
			delete(n.neighbors, d)
		}
	}
}

func (n *Neighbors) GetByBytes(d []byte) *Neighbor {
	n.lock.RLock()
	defer n.lock.RUnlock()

	k := binary.BigEndian.Uint32(d)
	if l, ok := n.neighbors[k]; ok {
		return l
	}
	return nil
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
	a.SetEther(a.pointCfg.If.Address)
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
	eth := a.NewEth(libol.ETHPARP, libol.BROADED)
	reply := libol.NewArp()
	reply.OpCode = libol.ARP_REQUEST
	reply.SIpAddr = a.Ether.IpAddr
	reply.TIpAddr = dest
	reply.SHwAddr = a.Ether.HwAddr
	reply.THwAddr = libol.ZEROED

	buffer := make([]byte, 0, a.pointCfg.If.Mtu)
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
			eth := a.NewEth(libol.ETHPIP4, neb.HwAddr)
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
		if arp.OpCode == libol.ARP_REQUEST && bytes.Equal(arp.TIpAddr, a.Ether.IpAddr) {
			eth := a.NewEth(libol.ETHPARP, arp.SHwAddr)
			reply := libol.NewArp()
			reply.OpCode = libol.ARP_REPLY
			reply.SIpAddr = a.Ether.IpAddr
			reply.TIpAddr = arp.SIpAddr
			reply.SHwAddr = a.Ether.HwAddr
			reply.THwAddr = arp.SHwAddr
			buffer := make([]byte, 0, a.pointCfg.If.Mtu)
			buffer = append(buffer, eth.Encode()...)
			buffer = append(buffer, reply.Encode()...)
			libol.Info("TapWorker.onArp: reply %x.", buffer)
			if a.Listener.ReadAt != nil {
				_ = a.Listener.ReadAt(buffer)
			}
		} else if arp.OpCode == libol.ARP_REPLY && bytes.Equal(arp.THwAddr, a.Ether.HwAddr) {
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
	if c.If.Provider == "tun" {
		return network.TapConfig{
			Type:    network.TUN,
			Name:    c.If.Name,
			Network: c.If.Address,
		}
	} else {
		return network.TapConfig{
			Type:    network.TAP,
			Name:    c.If.Name,
			Network: c.If.Address,
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
		IfAddr:      config.If.Address,
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
	p.tcpWorker = NewSessWorker(client, p.config)

	tapCfg := GetTapCfg(p.config)
	// register listener
	p.tapWorker = NewTapWorker(tapCfg, p.config)

	p.tcpWorker.SetUUID(p.UUID())
	p.tcpWorker.Listener = SessWorkerListener{
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
		nxt := net.ParseIP(rt.Nexthop)
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
