package _switch

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/switch/app"
	"github.com/danieldin95/openlan-go/src/switch/ctrls"
	"github.com/danieldin95/openlan-go/src/switch/storage"
	"strings"
	"sync"
	"time"
)

func GetSocketServer(c config.Switch) libol.SocketServer {
	switch c.Protocol {
	case "kcp":
		cfg := &libol.KcpConfig{
			Block:   config.GetBlock(c.Crypt),
			Timeout: time.Duration(c.Timeout) * time.Second,
		}
		return libol.NewKcpServer(c.Listen, cfg)
	case "tcp":
		cfg := &libol.TcpConfig{
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewTcpServer(c.Listen, cfg)
	case "udp":
		cfg := &libol.UdpConfig{
			Block:   config.GetBlock(c.Crypt),
			Timeout: time.Duration(c.Timeout) * time.Second,
		}
		return libol.NewUdpServer(c.Listen, cfg)
	case "ws":
		cfg := &libol.WebConfig{
			Block:   config.GetBlock(c.Crypt),
			Timeout: time.Duration(c.Timeout) * time.Second,
		}
		return libol.NewWebServer(c.Listen, cfg)
	case "wss":
		cfg := &libol.WebConfig{
			Ca: &libol.WebCa{
				CaCrt: c.Cert.CrtFile,
				CaKey: c.Cert.KeyFile,
			},
			Block:   config.GetBlock(c.Crypt),
			Timeout: time.Duration(c.Timeout) * time.Second,
		}
		return libol.NewWebServer(c.Listen, cfg)
	default:
		cfg := &libol.TcpConfig{
			Block: config.GetBlock(c.Crypt),
		}
		if c.Cert != nil {
			cfg.Tls = config.GetTlsCfg(*c.Cert)
		}
		return libol.NewTcpServer(c.Listen, cfg)
	}
}

type Apps struct {
	Auth     *app.PointAuth
	Request  *app.WithRequest
	Neighbor *app.Neighbors
	OnLines  *app.Online
}

type Hook func(client libol.SocketClient, frame *libol.FrameMessage) error

type Switch struct {
	// private
	lock     sync.Mutex
	cfg      config.Switch
	apps     Apps
	firewall FireWall
	hooks    []Hook
	http     *Http
	server   libol.SocketServer
	proxy    *Proxy
	worker   map[string]*NetworkWorker
	uuid     string
	newTime  int64
	out      *libol.SubLogger
}

func NewSwitch(c config.Switch) *Switch {
	server := GetSocketServer(c)
	v := Switch{
		cfg: c,
		firewall: FireWall{
			rules: make([]libol.IPTableRule, 0, 32),
		},
		worker:  make(map[string]*NetworkWorker, 32),
		server:  server,
		newTime: time.Now().Unix(),
		proxy:   NewProxy(c.Proxy),
		hooks:   make([]Hook, 0, 64),
		out:     libol.NewSubLogger(c.Alias),
	}
	return &v
}

func (v *Switch) acceptBridge(bridge string) {
	rules := v.firewall.rules
	v.out.Info("Switch.acceptBridge %s", bridge)
	rules = append(rules, libol.IPTableRule{
		Table: "filter",
		Chain: "FORWARD",
		Input: bridge,
	})
	v.firewall.rules = rules
}

func (v *Switch) acceptRoute(source, prefix string) {
	rules := v.firewall.rules
	v.out.Info("Switch.acceptRoute %s, %s", source, prefix)
	// allowed forward between source and prefix.
	rules = append(rules, libol.IPTableRule{
		Table:  "filter",
		Chain:  "FORWARD",
		Source: source,
		Dest:   prefix,
	})
	rules = append(rules, libol.IPTableRule{
		Table:  "filter",
		Chain:  "FORWARD",
		Source: prefix,
		Dest:   source,
	})
	// allowed input from source to prefix.
	rules = append(rules, libol.IPTableRule{
		Table:  "filter",
		Chain:  "INPUT",
		Source: source,
		Dest:   prefix,
	})
	// enable masquerade between source and prefix.
	rules = append(rules, libol.IPTableRule{
		Table:  "nat",
		Chain:  "POSTROUTING",
		Source: source,
		Dest:   prefix,
		Jump:   "MASQUERADE",
	})
	rules = append(rules, libol.IPTableRule{
		Table:  "nat",
		Chain:  "POSTROUTING",
		Source: prefix,
		Dest:   source,
		Jump:   "MASQUERADE",
	})
	v.firewall.rules = rules
}

func (v *Switch) initNetwork() {
	crypt := v.cfg.Crypt
	for _, nCfg := range v.cfg.Network {
		name := nCfg.Name
		brCfg := nCfg.Bridge
		// Forward traffic in bridge.
		v.acceptBridge(brCfg.Name)

		source := brCfg.Address
		ifAddr := strings.SplitN(source, "/", 2)[0]
		if ifAddr != "" {
			for _, rt := range nCfg.Routes {
				if rt.NextHop != ifAddr {
					continue
				}
				// MASQUERADE and allowed forward it.
				v.acceptRoute(source, rt.Prefix)
			}
		}
		v.worker[name] = NewNetworkWorker(*nCfg, crypt)
	}
}

func (v *Switch) initHook() {
	// Append accessed auth for point
	v.apps.Auth = app.NewPointAuth(v, v.cfg)
	v.hooks = append(v.hooks, v.apps.Auth.OnFrame)
	// Append request process
	v.apps.Request = app.NewWithRequest(v, v.cfg)
	v.hooks = append(v.hooks, v.apps.Request.OnFrame)
	// Check whether inspect neighbor
	if strings.Contains(v.cfg.Inspect, "neighbor") {
		v.apps.Neighbor = app.NewNeighbors(v, v.cfg)
		v.hooks = append(v.hooks, v.apps.Neighbor.OnFrame)
	}
	// Check whether inspect online flow by five-tuple.
	if strings.Contains(v.cfg.Inspect, "online") {
		v.apps.OnLines = app.NewOnline(v, v.cfg)
		v.hooks = append(v.hooks, v.apps.OnLines.OnFrame)
	}
	for i, h := range v.hooks {
		v.out.Debug("Switch.initHook: id %d, func %s", i, libol.FunName(h))
	}
}

func (v *Switch) initFirewall() {
	for _, rule := range v.cfg.FireWall {
		v.firewall.rules = append(v.firewall.rules, libol.IPTableRule{
			Table:    rule.Table,
			Chain:    rule.Chain,
			Source:   rule.Source,
			Dest:     rule.Dest,
			Jump:     rule.Jump,
			ToSource: rule.ToSource,
			ToDest:   rule.ToDest,
			Comment:  rule.Comment,
			Input:    rule.Input,
			Output:   rule.Output,
		})
	}
	v.out.Info("Switch.initFirewall total %d rules", len(v.firewall.rules))
}

func (v *Switch) initCtrl() {
	ctrls.Load(v.cfg.ConfDir + "/ctrl.json")
	if ctrls.Ctrl.Name == "" {
		ctrls.Ctrl.Name = v.cfg.Alias
	}
	ctrls.Ctrl.Switcher = v
}

func (v *Switch) Initialize() {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.initHook()
	if v.cfg.Http != nil {
		v.http = NewHttp(v, v.cfg)
	}
	v.initNetwork()
	// Controller
	v.initCtrl()
	// FireWall
	v.initFirewall()
	for _, w := range v.worker {
		w.Initialize()
	}
	v.proxy.Initialize()
}

func (v *Switch) onFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	for _, h := range v.hooks {
		if v.out.Has(libol.LOG) {
			v.out.Log("Switch.onFrame: %s", libol.FunName(h))
		}
		if h != nil {
			if err := h(client, frame); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *Switch) OnClient(client libol.SocketClient) error {
	client.SetStatus(libol.ClConnected)
	v.out.Info("Switch.onClient: %s", client.Address())
	return nil
}

func (v *Switch) SignIn(client libol.SocketClient) error {
	v.out.Cmd("Switch.SignIn %s", client.String())
	data := struct {
		Address string `json:"address"`
		Switch  string `json:"switch"`
	}{
		Address: client.String(),
		Switch:  client.LocalAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		v.out.Error("Switch.SignIn: %s", err)
		return err
	}
	v.out.Cmd("Switch.SignIn: %s", body)
	m := libol.NewControlFrame(libol.SignReq, body)
	if err := client.WriteMsg(m); err != nil {
		v.out.Error("Switch.SignIn: %s", err)
		return err
	}
	return nil
}

func (v *Switch) ReadClient(client libol.SocketClient, frame *libol.FrameMessage) error {
	addr := client.Address()
	if v.out.Has(libol.LOG) {
		v.out.Log("Switch.ReadClient: %s %x", addr, frame.Frame())
	}
	frame.Decode()
	if err := v.onFrame(client, frame); err != nil {
		v.out.Debug("Switch.ReadClient: %s dropping by %s", addr, err)
		if frame.Action() == libol.PingReq {
			// send sign message to point require login.
			_ = v.SignIn(client)
		}
		return nil
	}
	if frame.IsControl() {
		return nil
	}
	// process ethernet frame message.
	private := client.Private()
	if private != nil {
		point := private.(*models.Point)
		device := point.Device
		if point == nil || device == nil {
			return libol.NewErr("Tap devices is nil")
		}
		if _, err := device.Write(frame.Frame()); err != nil {
			v.out.Error("Switch.ReadClient: %s", err)
			return err
		}
		return nil
	}
	return libol.NewErr("point %s not found.", addr)
}

func (v *Switch) OnClose(client libol.SocketClient) error {
	addr := client.Address()
	v.out.Info("Switch.OnClose: %s", addr)
	// already not need support free list for device.
	uuid := storage.Point.GetUUID(addr)
	if storage.Point.GetAddr(uuid) == addr { // not has newer
		storage.Network.DelLease(uuid)
	}
	storage.Point.Del(addr)
	return nil
}

func (v *Switch) Start() {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.out.Debug("Switch.Start")
	if v.cfg.PProf != "" {
		f := libol.PProf{Listen: v.cfg.PProf}
		f.Start()
	}
	// firstly, start network.
	for _, w := range v.worker {
		w.Start(v)
	}
	// start server for accessing
	libol.Go(v.server.Accept)
	call := libol.ServerListener{
		OnClient: v.OnClient,
		OnClose:  v.OnClose,
		ReadAt:   v.ReadClient,
	}
	libol.Go(func() { v.server.Loop(call) })
	if v.http != nil {
		libol.Go(v.http.Start)
	}
	libol.Go(ctrls.Ctrl.Start)
	libol.Go(v.firewall.Start)
	if v.proxy != nil {
		v.proxy.Start()
	}
}

func (v *Switch) Stop() {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.out.Debug("Switch.Stop")
	if v.proxy != nil {
		v.proxy.Stop()
	}
	ctrls.Ctrl.Stop()
	// firstly, notify leave to point.
	for p := range storage.Point.List() {
		if p == nil {
			break
		}
		v.leftClient(p.Client)
	}
	v.firewall.Stop()
	if v.http != nil {
		v.http.Shutdown()
		v.http = nil
	}
	v.server.Close()
	// stop network.
	for _, w := range v.worker {
		w.Stop()
	}
}

func (v *Switch) Alias() string {
	return v.cfg.Alias
}

func (v *Switch) UpTime() int64 {
	return time.Now().Unix() - v.newTime
}

func (v *Switch) Server() libol.SocketServer {
	return v.server
}

func (v *Switch) NewTap(tenant string) (network.Taper, error) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.out.Debug("Switch.NewTap")

	// already not need support free list for device.
	// dropped firstly packages during 15s because of forwarding delay.
	w, ok := v.worker[tenant]
	if !ok {
		return nil, libol.NewErr("Not found bridge %s", tenant)
	}
	br := w.bridge
	dev, err := network.NewTaper(br.Type(), tenant, network.TapConfig{Type: network.TAP})
	if err != nil {
		v.out.Error("Switch.NewTap: %s", err)
		return nil, err
	}
	mtu := br.Mtu()
	dev.SetMtu(mtu)
	dev.Up()
	_ = br.AddSlave(dev)
	v.out.Info("Switch.NewTap: %s on %s", dev.Name(), tenant)
	return dev, nil
}

func (v *Switch) FreeTap(dev network.Taper) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	name := dev.Name()
	tenant := dev.Tenant()
	v.out.Debug("Switch.FreeTap %s", name)
	w, ok := v.worker[tenant]
	if !ok {
		return libol.NewErr("Not found bridge %s", tenant)
	}
	br := w.bridge
	_ = br.DelSlave(dev)
	v.out.Info("Switch.FreeTap: %s", name)
	return nil
}

func (v *Switch) UUID() string {
	if v.uuid == "" {
		v.uuid = libol.GenToken(32)
	}
	return v.uuid
}

func (v *Switch) AddLink(tenant string, c *config.Point) {
	//TODO dynamic configure
}

func (v *Switch) DelLink(tenant, addr string) {
	//TODO dynamic configure
}

func (v *Switch) ReadTap(device network.Taper, readAt func(f *libol.FrameMessage) error) {
	defer device.Close()
	name := device.Name()
	v.out.Info("Switch.ReadTap: %s", name)
	for {
		frame := libol.NewFrameMessage()
		n, err := device.Read(frame.Frame())
		if err != nil {
			v.out.Error("Switch.ReadTap: %s", err)
			break
		}
		frame.SetSize(n)
		if v.out.Has(libol.LOG) {
			v.out.Log("Switch.ReadTap: %x\n", frame.Frame()[:n])
		}
		if err := readAt(frame); err != nil {
			v.out.Error("Switch.ReadTap: readAt %s %s", name, err)
			break
		}
	}
}

func (v *Switch) OffClient(client libol.SocketClient) {
	v.out.Info("Switch.OffClient: %s", client)
	if v.server != nil {
		v.server.OffClient(client)
	}
}

func (v *Switch) Config() *config.Switch {
	return &v.cfg
}

func (v *Switch) leftClient(client libol.SocketClient) {
	if client == nil {
		return
	}
	v.out.Info("Switch.leftClient: %s", client.String())
	data := struct {
		DateTime   int64  `json:"datetime"`
		UUID       string `json:"uuid"`
		Alias      string `json:"alias"`
		Connection string `json:"connection"`
		Address    string `json:"address"`
	}{
		DateTime:   time.Now().Unix(),
		UUID:       v.UUID(),
		Alias:      v.Alias(),
		Address:    client.LocalAddr(),
		Connection: client.RemoteAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		v.out.Error("Switch.leftClient: %s", err)
		return
	}
	v.out.Cmd("Switch.leftClient: %s", body)
	m := libol.NewControlFrame(libol.LeftReq, body)
	if err := client.WriteMsg(m); err != nil {
		v.out.Error("Switch.leftClient: %s", err)
		return
	}
}
