package _switch

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/network"
	"github.com/danieldin95/openlan-go/switch/app"
	"github.com/danieldin95/openlan-go/switch/ctrls"
	"github.com/danieldin95/openlan-go/switch/storage"
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
	bridge   map[string]network.Bridger
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
		bridge:  make(map[string]network.Bridger, 32),
		server:  server,
		newTime: time.Now().Unix(),
	}
	return &v
}

func (v *Switch) addRules(source string, prefix string) {
	libol.Info("Switch.addRules %s, %s", source, prefix)
	v.firewall.rules = append(v.firewall.rules, libol.FilterRule{
		Table:  "filter",
		Chain:  "FORWARD",
		Source: source,
		Dest:   prefix,
		Jump:   "ACCEPT",
	})
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
		Dest:   source,
		Source: prefix,
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

		if brCfg.Address != "" {
			source := brCfg.Address
			ifAddr := strings.SplitN(source, "/", 2)[0]
			for i, rt := range nCfg.Routes {
				if rt.NextHop == "" {
					nCfg.Routes[i].NextHop = ifAddr
				}
				rt = nCfg.Routes[i]
				if rt.NextHop != ifAddr {
					continue
				}
				// MASQUERADE
				v.addRules(source, rt.Prefix)
			}
		}
		v.worker[name] = NewNetworkWorker(*nCfg, crypt)
		v.bridge[name] = network.NewBridger(brCfg.Provider, brCfg.Name, brCfg.Mtu)
	}

	v.apps.Auth = app.NewPointAuth(v, v.cfg)
	v.apps.Request = app.NewWithRequest(v, v.cfg)
	v.apps.Neighbor = app.NewNeighbors(v, v.cfg)
	v.apps.OnLines = app.NewOnline(v, v.cfg)

	v.hooks = make([]Hook, 0, 64)
	v.hooks = append(v.hooks, v.apps.Auth.OnFrame)
	v.hooks = append(v.hooks, v.apps.Neighbor.OnFrame)
	v.hooks = append(v.hooks, v.apps.Request.OnFrame)
	v.hooks = append(v.hooks, v.apps.OnLines.OnFrame)
	for i, h := range v.hooks {
		libol.Debug("Switch.Initialize: k %d, func %p, %s", i, h, libol.FunName(h))
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
}

func (v *Switch) onFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	for _, h := range v.hooks {
		libol.Log("Switch.onFrame: h %p", h)
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

func (v *Switch) ReadClient(client libol.SocketClient, data []byte) error {
	libol.Log("Switch.ReadClient: %s %x", client.Addr(), data)
	frame := libol.NewFrameMessage(data)
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
		dev := point.Device
		if point == nil || dev == nil {
			return libol.NewErr("Tap devices is nil")
		}
		if _, err := dev.Write(data); err != nil {
			libol.Error("Switch.ReadClient: %s", err)
			return err
		}
		return nil
	}
	return libol.NewErr("point %s not found.", client)
}

func (v *Switch) OnClose(client libol.SocketClient) error {
	libol.Info("Switch.OnClose: %s", client.Addr())

	uuid := storage.Point.GetUUID(client.Addr())
	if storage.Point.GetAddr(uuid) == client.Addr() { // not has newer
		storage.Network.FreeAddr(uuid)
	}
	storage.Point.Del(client.Addr())

	return nil
}

func (v *Switch) Start() {
	v.lock.Lock()
	defer v.lock.Unlock()

	libol.Debug("Switch.Start")
	for _, nCfg := range v.cfg.Network {
		if br, ok := v.bridge[nCfg.Name]; ok {
			brCfg := nCfg.Bridge
			br.Open(brCfg.Address)
		}
	}
	libol.Go(v.server.Accept)
	call := libol.ServerListener{
		OnClient: v.OnClient,
		OnClose:  v.OnClose,
		ReadAt:   v.ReadClient,
	}
	libol.Go(func() {v.server.Loop(call)})
	for _, w := range v.worker {
		w.Start(v)
	}
	if v.http != nil {
		libol.Go(v.http.Start)
	}
	libol.Go(ctrls.Ctrl.Start)
	libol.Go(v.firewall.Start)
}

func (v *Switch) Stop() {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.bridge == nil {
		return
	}
	libol.Debug("Switch.Stop")
	for p := range storage.Point.List() {
		if p == nil {
			break
		}
		v.leftClient(p.Client)
	}
	v.firewall.Stop()
	ctrls.Ctrl.Stop()
	if v.http != nil {
		v.http.Shutdown()
		v.http = nil
	}
	for _, w := range v.worker {
		w.Stop()
	}
	for _, nCfg := range v.cfg.Network {
		if br, ok := v.bridge[nCfg.Name]; ok {
			brCfg := nCfg.Bridge
			_ = br.Close()
			delete(v.bridge, brCfg.Name)
		}
	}
	v.server.Close()
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

	br, ok := v.bridge[tenant]
	if !ok {
		return nil, libol.NewErr("Not found bridge %s", tenant)
	}
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
	br, ok := v.bridge[dev.Tenant()]
	if !ok {
		return libol.NewErr("Not found bridge %s", dev.Tenant())
	}
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
	//TODO
}

func (v *Switch) DelLink(tenant, addr string) {
	//TODO
}

func (v *Switch) ReadTap(dev network.Taper, readAt func(p []byte) error) {
	defer dev.Close()
	libol.Info("Switch.ReadTap: %s", dev.Name())

	data := make([]byte, libol.MAXBUF)
	for {
		n, err := dev.Read(data)
		if err != nil {
			libol.Error("Switch.ReadTap: %s", err)
			break
		}
		libol.Log("Switch.ReadTap: %x\n", data[:n])
		if err := readAt(data[:n]); err != nil {
			libol.Error("Switch.ReadTap: do-recv %s %s", dev.Name(), err)
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
