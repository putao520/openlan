package point

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/network"
	"github.com/songgao/water"
	"net"
	"strings"
	"sync"
	"time"
)

type TcpWorkerListener struct {
	OnClose   func(w *TcpWorker) error
	OnSuccess func(w *TcpWorker) error
	OnIpAddr  func(w *TcpWorker, n *models.Network) error
	ReadAt    func(p []byte) error
}

type TcpWorker struct {
	lock     sync.RWMutex
	Listener TcpWorkerListener
	Client   *libol.TcpClient

	writeChan   chan []byte
	maxSize     int
	alias       string
	user        *models.User
	network     *models.Network
	routes      map[string]*models.Route
	allowed     bool
	initialized bool
}

func NewTcpWorker(client *libol.TcpClient, c *config.Point) (t *TcpWorker) {
	t = &TcpWorker{
		Client:      client,
		writeChan:   make(chan []byte, 1024*10),
		maxSize:     c.IfMtu,
		user:        models.NewUser(c.Name(), c.Password()),
		network:     models.NewNetwork(c.Tenant, c.IfAddr),
		routes:      make(map[string]*models.Route, 64),
		allowed:     c.Allowed,
		initialized: false,
	}
	t.user.Alias = c.Alias
	t.user.Tenant = c.Tenant

	return
}

func (t *TcpWorker) Initialize() {
	if t.Client == nil {
		return
	}

	libol.Info("TcpWorker.Initialize")
	t.initialized = true
	t.Client.SetMaxSize(t.maxSize)
	t.Client.Listener = libol.TcpClientListener{
		OnConnected: func(client *libol.TcpClient) error {
			return t.TryLogin(client)
		},
		OnClose: func(client *libol.TcpClient) error {
			if t.Listener.OnClose != nil {
				_ = t.Listener.OnClose(t)
			}
			return nil
		},
	}
}

func (t *TcpWorker) Start() {
	if !t.initialized {
		t.Initialize()
	}
	_ = t.Connect()

	go t.Read()
	go t.Loop()
}

func (t *TcpWorker) Stop() {
	t.Client.Terminal()

	t.lock.Lock()
	defer t.lock.Unlock()

	close(t.writeChan)
	t.Client = nil
}

func (t *TcpWorker) Close() {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.Client == nil {
		return
	}
	t.Client.Close()
}

func (t *TcpWorker) Connect() error {
	s := t.Client.Status()
	if s != libol.CL_INIT {
		libol.Warn("TcpWorker.Connect status %d->%d", s, libol.CL_INIT)
		t.Client.SetStatus(libol.CL_INIT)
	}

	if err := t.Client.Connect(); err != nil {
		libol.Error("TcpWorker.Connect %s", err)
		return err
	}
	return nil
}

func (t *TcpWorker) TryLogin(client *libol.TcpClient) error {
	body, err := json.Marshal(t.user)
	if err != nil {
		libol.Error("TcpWorker.TryLogin: %s", err)
		return err
	}

	libol.Info("TcpWorker.TryLogin: %s", body)
	if err := client.WriteReq("login", string(body)); err != nil {
		return err
	}
	return nil
}

func (t *TcpWorker) TryNetwork(client *libol.TcpClient) error {
	body, err := json.Marshal(t.network)
	if err != nil {
		libol.Error("TcpWorker.TryNetwork: %s", err)
		return err
	}

	libol.Info("TcpWorker.TryNetwork: %s", body)
	if err := client.WriteReq("ipaddr", string(body)); err != nil {
		return err
	}
	return nil
}

func (t *TcpWorker) onInstruct(data []byte) error {
	m := libol.NewFrameMessage(data)
	if !m.IsControl() {
		return nil
	}

	action, resp := m.CmdAndParams()
	if action == "logi:" {
		libol.Debug("TcpWorker.onInstruct.login: %s", resp)
		if resp[:4] == "okay" {
			t.Client.SetStatus(libol.CL_AUEHED)
			if t.Listener.OnSuccess != nil {
				_ = t.Listener.OnSuccess(t)
			}
			if t.allowed {
				_ = t.TryNetwork(t.Client)
			}
			libol.Info("TcpWorker.onInstruct.login: success")
		} else {
			t.Client.SetStatus(libol.CL_UNAUTH)
			libol.Error("TcpWorker.onInstruct.login: %s", resp)
		}

		return nil
	}

	if libol.IsErrorResponse(resp) {
		libol.Error("TcpWorker.onInstruct.%s: %s", action, resp)
		return nil
	}

	if action == "ipad:" {
		net := models.Network{}
		if err := json.Unmarshal([]byte(resp), &net); err != nil {
			return libol.NewErr("TcpWorker.onInstruct: Invalid json data.")
		}

		libol.Debug("TcpWorker.onInstruct.ipaddr: %s", resp)
		if t.Listener.OnIpAddr != nil {
			_ = t.Listener.OnIpAddr(t, &net)
		}

	}

	return nil
}

func (t *TcpWorker) Read() {
	libol.Info("TcpWorker.Read %t", t.Client.IsOk())
	defer libol.Catch("TcpWorker.Read")

	for {
		if t.Client == nil || t.Client.IsTerminal() {
			break
		}

		if !t.Client.IsOk() {
			time.Sleep(30 * time.Second) // sleep 30s and release cpu.
			_ = t.Connect()
			continue
		}

		data := make([]byte, t.maxSize)
		n, err := t.Client.ReadMsg(data)
		if err != nil {
			libol.Error("TcpWorker.Read: %s", err)
			t.Close()
			continue
		}

		libol.Debug("TcpWorker.Read: %x", data[:n])
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
	libol.Info("TcpWorker.Read exit")
}

func (t *TcpWorker) DoWrite(data []byte) error {
	libol.Debug("TcpWorker.DoWrite: %x", data)

	t.writeChan <- data

	return nil
}

func (t *TcpWorker) Loop() {
	defer libol.Catch("TcpWorker.Loop")

	for {
		w, ok := <-t.writeChan
		if !ok || t.Client == nil {
			break
		}

		if t.Client.Status() != libol.CL_AUEHED {
			t.Client.Sts.Dropped++
			libol.Debug("TcpWorker.Loop: dropping by unAuth")
			continue
		}
		if err := t.Client.WriteMsg(w); err != nil {
			libol.Error("TcpWorker.Loop: %s", err)
			break
		}
	}

	t.Close()
	libol.Info("TcpWorker.Loop exit")
}

func (t *TcpWorker) Auth() (string, string) {
	return t.user.Name, t.user.Password
}

func (t *TcpWorker) SetAuth(auth string) {
	values := strings.Split(auth, ":")
	t.user.Name = values[0]
	if len(values) > 1 {
		t.user.Password = values[1]
	}
}

func (t *TcpWorker) SetAddr(addr string) {
	t.Client.Addr = addr
}

func (t *TcpWorker) SetUUID(v string) {
	t.user.UUID = v
}

type TapWorkerListener struct {
	OnOpen  func(w *TapWorker) error
	OnClose func(w *TapWorker)
	ReadAt  func([]byte) error
}

type TapWorker struct {
	Device   network.Taper
	Listener TapWorkerListener
	//for tunnel device.
	EthDstAddr []byte
	EthSrcAddr []byte
	EthSrcIp   []byte

	lock        sync.RWMutex
	writeChan   chan []byte
	devCfg      *water.Config
	pointCfg    *config.Point
	initialized bool
}

func NewTapWorker(devCfg *water.Config, c *config.Point) (a *TapWorker) {
	a = &TapWorker{
		Device:      nil,
		devCfg:      devCfg,
		pointCfg:    c,
		writeChan:   make(chan []byte, 1024*10),
		initialized: false,
	}

	return
}

func (a *TapWorker) Initialize() {
	a.initialized = true
	libol.Info("TapWorker.Initialize")
	a.Open()
	a.DoTun()
}

func (a *TapWorker) DoTun() {
	if a.Device == nil || !a.Device.IsTun() {
		return
	}

	a.EthSrcIp = net.ParseIP(a.pointCfg.IfAddr).To4()
	libol.Info("NewTapWorker srcIp: % x", a.EthSrcIp)

	if a.pointCfg.IfEthSrc == "" {
		a.EthSrcAddr = libol.GenEthAddr(6)
	} else {
		if hw, err := net.ParseMAC(a.pointCfg.IfEthSrc); err == nil {
			a.EthSrcAddr = hw
		}
	}
	if hw, err := net.ParseMAC(a.pointCfg.IfEthDst); err == nil {
		a.EthDstAddr = hw
	}
	libol.Info("NewTapWorker src: %x, dst: %x", a.EthSrcAddr, a.EthDstAddr)

}

func (a *TapWorker) Open() {
	if a.Device != nil {
		_ = a.Device.Close()
		time.Sleep(5 * time.Second) // sleep 5s and release cpu.
	}

	var err error
	var dev network.Taper
	if a.devCfg.DeviceType == water.TAP {
		dev, err = network.NewKernelTap(true, a.pointCfg.Tenant, "")
	} else {
		dev, err = network.NewKernelTap(false, a.pointCfg.Tenant, "")
	}
	if err != nil {
		libol.Error("TapWorker.Open %s", err)
		return
	}

	libol.Info("TapWorker.Open %s", dev.Name())
	a.Device = dev
	if a.Listener.OnOpen != nil {
		_ = a.Listener.OnOpen(a)
	}
}

func (a *TapWorker) NewEth(t uint16) *libol.Ether {
	eth := libol.NewEther(t)
	eth.Dst = a.EthDstAddr
	eth.Src = a.EthSrcAddr

	return eth
}

func (a *TapWorker) Read() {
	defer libol.Catch("TapWorker.Read")

	libol.Info("TapWorker.Read")
	for {
		if a.Device == nil {
			break
		}

		data := make([]byte, a.pointCfg.IfMtu)
		n, err := a.Device.Read(data)
		if err != nil {
			libol.Error("TapWorker.Read: %s", err)
			a.Open()
			continue
		}

		libol.Debug("TapWorker.Read: %x", data[:n])
		if a.Device.IsTun() {
			eth := a.NewEth(libol.ETHPIP4)

			buffer := make([]byte, 0, a.pointCfg.IfMtu)
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
	libol.Info("TapWorker.Read exit")
}

func (a *TapWorker) DoWrite(data []byte) error {
	libol.Debug("TapWorker.DoWrite: %x", data)

	a.writeChan <- data

	return nil
}

func (a *TapWorker) onArp(data []byte) bool {
	libol.Debug("TapWorker.onArp")
	eth, err := libol.NewEtherFromFrame(data)
	if err != nil {
		libol.Warn("TapWorker.onArp %s", err)
		return false
	}

	if !eth.IsArp() {
		return false
	}

	arp, err := libol.NewArpFromFrame(data[eth.Len:])
	if err != nil {
		libol.Error("TapWorker.onArp %s.", err)
		return false
	}

	if arp.IsIP4() {
		if arp.OpCode != libol.ARP_REQUEST {
			return false
		}

		eth := a.NewEth(libol.ETHPARP)

		reply := libol.NewArp()
		reply.OpCode = libol.ARP_REPLY
		reply.SIpAddr = a.EthSrcIp
		reply.TIpAddr = arp.SIpAddr
		reply.SHwAddr = a.EthSrcAddr
		reply.THwAddr = arp.SHwAddr

		buffer := make([]byte, 0, a.pointCfg.IfMtu)
		buffer = append(buffer, eth.Encode()...)
		buffer = append(buffer, reply.Encode()...)

		libol.Info("TapWorker.onArp %x.", buffer)
		if a.Listener.ReadAt != nil {
			_ = a.Listener.ReadAt(buffer)
		}
		return true
	}
	return false
}

func (a *TapWorker) Loop() {
	libol.Info("TapWorker.Loop")
	defer libol.Catch("TapWorker.Loop")

	for {
		w, ok := <-a.writeChan
		if a.Device == nil || !ok {
			break
		}

		if a.Device.IsTun() {
			//Proxy arp request.
			if a.onArp(w) {
				libol.Info("TapWorker.Loop: Arp proxy.")
				continue
			}

			eth, err := libol.NewEtherFromFrame(w)
			if err != nil {
				libol.Error("TapWorker.Loop: %s", err)
				continue
			}
			if eth.IsVlan() {
				w = w[18:]
			} else if eth.IsIP4() {
				w = w[14:]
			} else { // default is Ethernet is 14 bytes.
				w = w[14:]
			}
		}

		if _, err := a.Device.Write(w); err != nil {
			libol.Error("TapWorker.Loop: %s", err)
		}
	}

	a.Close()
	libol.Info("TapWorker.Loop exit")
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
	go a.Loop()
}

func (a *TapWorker) Stop() {
	close(a.writeChan)
	a.Close()
}

type WorkerListener struct {
	AddAddr   func(ipStr string) error
	DelAddr   func(ipStr string) error
	OnTap     func(w *TapWorker) error
	AddRoutes func(routes []*models.Route) error
	DelRoutes func(routes []*models.Route) error
}

type Worker struct {
	IfAddr   string
	Listener WorkerListener

	tcpWorker   *TcpWorker
	tapWorker   *TapWorker
	config      *config.Point
	uuid        string
	network     *models.Network
	initialized bool
}

func NewWorker(config *config.Point) (p *Worker) {
	p = &Worker{
		IfAddr:      config.IfAddr,
		config:      config,
		initialized: false,
	}

	return
}

func (p *Worker) Initialize() {
	var conf *water.Config
	var tlsConf *tls.Config

	if p.config == nil {
		return
	}

	libol.Info("Worker.Initialize")

	p.initialized = true
	if p.config.Tls {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}
	client := libol.NewTcpClient(p.config.Addr, tlsConf)
	p.tcpWorker = NewTcpWorker(client, p.config)

	if p.config.IfTun {
		conf = &water.Config{DeviceType: water.TUN}
	} else {
		conf = &water.Config{DeviceType: water.TAP}
	}

	// register listener
	p.tapWorker = NewTapWorker(conf, p.config)

	p.tcpWorker.SetUUID(p.UUID())
	p.tcpWorker.Listener = TcpWorkerListener{
		OnClose:   p.OnClose,
		OnSuccess: p.OnSuccess,
		OnIpAddr:  p.OnIpAddr,
		ReadAt:    p.tapWorker.DoWrite,
	}
	p.tcpWorker.Initialize()

	p.tapWorker.Listener = TapWorkerListener{
		OnOpen: p.Listener.OnTap,
		ReadAt: p.tcpWorker.DoWrite,
	}
	p.tapWorker.Initialize()
}

func (p *Worker) Start() {
	libol.Debug("Worker.Start linux.")
	if !p.initialized {
		p.Initialize()
	}
	p.tapWorker.Start()
	p.tcpWorker.Start()
}

func (p *Worker) Stop() {
	if p.tapWorker == nil || p.tcpWorker == nil {
		return
	}

	p.FreeIpAddr()
	p.tcpWorker.Stop()
	p.tapWorker.Stop()
	p.tcpWorker = nil
	p.tapWorker = nil
}

func (p *Worker) Client() *libol.TcpClient {
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
		return client.Addr
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

func (p *Worker) Worker() *TcpWorker {
	if p.tcpWorker != nil {
		return p.tcpWorker
	}
	return nil
}

func (p *Worker) OnIpAddr(w *TcpWorker, n *models.Network) error {
	libol.Info("Worker.OnIpAddr: %s/%s, %s", n.IfAddr, n.Netmask, n.Routes)

	prefix := libol.Netmask2Len(n.Netmask)
	ipStr := fmt.Sprintf("%s/%d", n.IfAddr, prefix)

	if p.Listener.AddAddr != nil {
		_ = p.Listener.AddAddr(ipStr)
	}
	if p.Listener.AddRoutes != nil {
		_ = p.Listener.AddRoutes(n.Routes)
	}
	p.network = n

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

func (p *Worker) OnClose(w *TcpWorker) error {
	libol.Info("Worker.OnClose")
	p.FreeIpAddr()
	return nil
}

func (p *Worker) OnSuccess(w *TcpWorker) error {
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
