package _switch

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/cli/config"
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
		kcpCfg := &libol.KcpConfig{
			Block:   config.GetBlock(c.Crypt),
			Timeout: time.Duration(c.Timeout) * time.Second,
		}
		return libol.NewKcpServer(c.Listen, kcpCfg)
	case "tcp":
		tcpCfg := &libol.TcpConfig{
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewTcpServer(c.Listen, tcpCfg)
	case "udp":
		udpCfg := &libol.UdpConfig{
			Block:   config.GetBlock(c.Crypt),
			Timeout: time.Duration(c.Timeout) * time.Second,
		}
		return libol.NewUdpServer(c.Listen, udpCfg)
	default:
		tcpCfg := &libol.TcpConfig{
			Tls:   config.GetTlsCfg(c.Cert),
			Block: config.GetBlock(c.Crypt),
		}
		return libol.NewTcpServer(c.Listen, tcpCfg)
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
	worker   map[string]*NetworkWorker
	uuid     string
	newTime  int64
}

func NewSwitch(c config.Switch) *Switch {
	server := GetSocketServer(c)
	v := Switch{
		cfg: c,
		firewall: FireWall{
			rules: make([]libol.FilterRule, 0, 32),
		},
		worker:  make(map[string]*NetworkWorker, 32),
		server:  server,
		newTime: time.Now().Unix(),
	}
	return &v
}

func (v *Switch) addRules(source string, prefix string) {
	libol.Info("Switch.addRules %s, %s", source, prefix)
	// allowed between source and prefix on filter.
	v.firewall.rules = append(v.firewall.rules, libol.FilterRule{
		Table:  "filter",
		Chain:  "FORWARD",
		Source: source,
		Dest:   prefix,
		Jump:   "ACCEPT",
	})
	v.firewall.rules = append(v.firewall.rules, libol.FilterRule{
		Table:  "filter",
		Chain:  "FORWARD",
		Source: prefix,
		Dest:   source,
		Jump:   "ACCEPT",
	})
	// enable masquerade between source and prefix.
	v.firewall.rules = append(v.firewall.rules, libol.FilterRule{
		Table:  "nat",
		Chain:  "POSTROUTING",
		Source: source,
		Dest:   prefix,
		Jump:   "MASQUERADE",
	})
	v.firewall.rules = append(v.firewall.rules, libol.FilterRule{
		Table:  "nat",
		Chain:  "POSTROUTING",
		Source: prefix,
		Dest:   source,
		Jump:   "MASQUERADE",
	})
}

func (v *Switch) Initialize() {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.cfg.Http != nil {
		v.http = NewHttp(v, v.cfg)
	}
	crypt := v.cfg.Crypt
	for _, nCfg := range v.cfg.Network {
		name := nCfg.Name
		brCfg := nCfg.Bridge

		source := brCfg.Address
		ifAddr := strings.SplitN(source, "/", 2)[0]
		if ifAddr != "" {
			for _, rt := range nCfg.Routes {
				if rt.NextHop != ifAddr {
					continue
				}
				// MASQUERADE
				v.addRules(source, rt.Prefix)
			}
		}
		v.worker[name] = NewNetworkWorker(*nCfg, crypt)
	}

	v.hooks = make([]Hook, 0, 64)
	v.apps.Auth = app.NewPointAuth(v, v.cfg)
	v.hooks = append(v.hooks, v.apps.Auth.OnFrame)
	v.apps.Request = app.NewWithRequest(v, v.cfg)
	v.hooks = append(v.hooks, v.apps.Request.OnFrame)
	if strings.Contains(v.cfg.Inspect, "neighbor") {
		v.apps.Neighbor = app.NewNeighbors(v, v.cfg)
		v.hooks = append(v.hooks, v.apps.Neighbor.OnFrame)
	}
	if strings.Contains(v.cfg.Inspect, "online") {
		v.apps.OnLines = app.NewOnline(v, v.cfg)
		v.hooks = append(v.hooks, v.apps.OnLines.OnFrame)
	}
	for i, h := range v.hooks {
		libol.Info("Switch.Initialize: id %d, func %s", i, libol.FunName(h))
	}

	// Controller
	ctrls.Load(v.cfg.ConfDir + "/ctrl.json")
	if ctrls.Ctrl.Name == "" {
		ctrls.Ctrl.Name = v.cfg.Alias
	}
	ctrls.Ctrl.Switcher = v

	// FireWall
	for _, rule := range v.cfg.FireWall {
		v.firewall.rules = append(v.firewall.rules, libol.FilterRule{
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
	libol.Info("Switch.Initialize total %d rules", len(v.firewall.rules))
	for _, w := range v.worker {
		w.Initialize()
	}
}

func (v *Switch) onFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	for _, h := range v.hooks {
		if libol.HasLog(libol.LOG) {
			libol.Log("Switch.onFrame: %s", libol.FunName(h))
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
	libol.Info("Switch.onClient: %s", client.Addr())
	return nil
}

func (v *Switch) SignIn(client libol.SocketClient) error {
	libol.Cmd("Switch.SignIn %s", client.String())
	data := struct {
		Address string `json:"address"`
		Switch  string `json:"switch"`
	}{
		Address: client.String(),
		Switch:  client.LocalAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		libol.Error("Switch.SignIn: %s", err)
		return err
	}
	libol.Cmd("Switch.SignIn: %s", body)
	if err := client.WriteReq("signin", string(body)); err != nil {
		libol.Error("Switch.SignIn: %s", err)
		return err
	}
	return nil
}

func (v *Switch) ReadClient(client libol.SocketClient, frame *libol.FrameMessage) error {
	if libol.HasLog(libol.LOG) {
		libol.Log("Switch.ReadClient: %s %x", client.Addr(), frame.Frame())
	}
	frame.Decode()
	if err := v.onFrame(client, frame); err != nil {
		libol.Debug("Switch.ReadClient: %s dropping by %s", client.Addr(), err)
		// send request to point login again.
		_ = v.SignIn(client)
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
			libol.Error("Switch.ReadClient: %s", err)
			return err
		}
		return nil
	}
	return libol.NewErr("point %s not found.", client)
}

func (v *Switch) OnClose(client libol.SocketClient) error {
	libol.Info("Switch.OnClose: %s", client.Addr())

	// already not need support free list for device.
	uuid := storage.Point.GetUUID(client.Addr())
	if storage.Point.GetAddr(uuid) == client.Addr() { // not has newer
		storage.Network.DelUsedAddr(uuid)
	}
	storage.Point.Del(client.Addr())

	return nil
}

func (v *Switch) Start() {
	v.lock.Lock()
	defer v.lock.Unlock()

	libol.Debug("Switch.Start")
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
}

func (v *Switch) Stop() {
	v.lock.Lock()
	defer v.lock.Unlock()

	libol.Debug("Switch.Stop")
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
	libol.Debug("Switch.NewTap")

	// already not need support free list for device.
	// dropped firstly packages during 15s because of forwarding delay.
	w, ok := v.worker[tenant]
	if !ok {
		return nil, libol.NewErr("Not found bridge %s", tenant)
	}
	br := w.bridge
	dev, err := network.NewTaper(br.Type(), tenant, network.TapConfig{Type: network.TAP})
	if err != nil {
		libol.Error("Switch.NewTap: %s", err)
		return nil, err
	}
	mtu := br.Mtu()
	dev.SetMtu(mtu)
	dev.Up()
	_ = br.AddSlave(dev)
	libol.Info("Switch.NewTap: %s on %s", dev.Name(), tenant)
	return dev, nil
}

func (v *Switch) FreeTap(dev network.Taper) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	libol.Debug("Switch.FreeTap %s", dev.Name())

	w, ok := v.worker[dev.Tenant()]
	if !ok {
		return libol.NewErr("Not found bridge %s", dev.Tenant())
	}
	br := w.bridge
	_ = br.DelSlave(dev)
	libol.Info("Switch.FreeTap: %s", dev.Name())
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
	libol.Info("Switch.ReadTap: %s", device.Name())

	for {
		frame := libol.NewFrameMessage()
		n, err := device.Read(frame.Frame())
		if err != nil {
			libol.Error("Switch.ReadTap: %s", err)
			break
		}
		frame.SetSize(n)
		if libol.HasLog(libol.LOG) {
			libol.Log("Switch.ReadTap: %x\n", frame.Frame()[:n])
		}
		if err := readAt(frame); err != nil {
			libol.Error("Switch.ReadTap: do-recv %s %s", device.Name(), err)
			break
		}
	}
}

func (v *Switch) OffClient(client libol.SocketClient) {
	libol.Info("Switch.OffClient: %s", client)
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
	libol.Info("Switch.leftClient: %s", client.String())
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
		libol.Error("Switch.leftClient: %s", err)
		return
	}
	libol.Cmd("Switch.leftClient: %s", body)
	if err := client.WriteReq("left", string(body)); err != nil {
		libol.Error("Switch.leftClient: %s", err)
		return
	}
}
