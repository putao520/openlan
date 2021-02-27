package olsw

import (
	"bufio"
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olsw/app"
	"github.com/danieldin95/openlan-go/src/olsw/ctrls"
	"github.com/danieldin95/openlan-go/src/olsw/store"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func GetSocketServer(s *config.Switch) libol.SocketServer {
	switch s.Protocol {
	case "kcp":
		c := &libol.KcpConfig{
			Block:   config.GetBlock(s.Crypt),
			Timeout: time.Duration(s.Timeout) * time.Second,
		}
		return libol.NewKcpServer(s.Listen, c)
	case "tcp":
		c := &libol.TcpConfig{
			Block:   config.GetBlock(s.Crypt),
			Timeout: time.Duration(s.Timeout) * time.Second,
			RdQus:   s.Queue.SockRd,
			WrQus:   s.Queue.SockWr,
		}
		return libol.NewTcpServer(s.Listen, c)
	case "udp":
		c := &libol.UdpConfig{
			Block:   config.GetBlock(s.Crypt),
			Timeout: time.Duration(s.Timeout) * time.Second,
		}
		return libol.NewUdpServer(s.Listen, c)
	case "ws":
		c := &libol.WebConfig{
			Block:   config.GetBlock(s.Crypt),
			Timeout: time.Duration(s.Timeout) * time.Second,
			RdQus:   s.Queue.SockRd,
			WrQus:   s.Queue.SockWr,
		}
		return libol.NewWebServer(s.Listen, c)
	case "wss":
		c := &libol.WebConfig{
			Block:   config.GetBlock(s.Crypt),
			Timeout: time.Duration(s.Timeout) * time.Second,
			RdQus:   s.Queue.SockRd,
			WrQus:   s.Queue.SockWr,
		}
		if s.Cert != nil {
			c.Cert = &libol.WebCert{
				Crt: s.Cert.CrtFile,
				Key: s.Cert.KeyFile,
			}
		}
		return libol.NewWebServer(s.Listen, c)
	default:
		c := &libol.TcpConfig{
			Block:   config.GetBlock(s.Crypt),
			Timeout: time.Duration(s.Timeout) * time.Second,
			RdQus:   s.Queue.SockRd,
			WrQus:   s.Queue.SockWr,
		}
		if s.Cert != nil {
			c.Tls = s.Cert.GetTlsCfg()
		}
		return libol.NewTcpServer(s.Listen, c)
	}
}

type Apps struct {
	Auth     *app.Access
	Request  *app.Request
	Neighbor *app.Neighbors
	OnLines  *app.Online
}

type Hook func(client libol.SocketClient, frame *libol.FrameMessage) error

type Switch struct {
	// private
	lock     sync.Mutex
	cfg      *config.Switch
	apps     Apps
	firewall *FireWall
	hooks    []Hook
	http     *Http
	server   libol.SocketServer
	worker   map[string]*NetworkWorker
	uuid     string
	newTime  int64
	out      *libol.SubLogger
}

func NewSwitch(c *config.Switch) *Switch {
	server := GetSocketServer(c)
	v := Switch{
		cfg:      c,
		firewall: NewFireWall(c.FireWall),
		worker:   make(map[string]*NetworkWorker, 32),
		server:   server,
		newTime:  time.Now().Unix(),
		hooks:    make([]Hook, 0, 64),
		out:      libol.NewSubLogger(c.Alias),
	}
	return &v
}

func (v *Switch) allowForward(bridge, source, prefix string) {
	if source == prefix {
		return
	}
	v.out.Info("Switch.allowForward %s, %s", source, prefix)
	// allowed forward between source and prefix.
	v.firewall.AddRule(libol.IptRule{
		Table:  FilterT,
		Chain:  OlForwardC,
		Input:  bridge,
		Source: source,
		Dest:   prefix,
	})
	v.firewall.AddRule(libol.IptRule{
		Table:  FilterT,
		Chain:  OlForwardC,
		Output: bridge,
		Source: prefix,
		Dest:   source,
	})
	// allowed input from source to prefix.
	v.firewall.AddRule(libol.IptRule{
		Table:  FilterT,
		Chain:  OlInputC,
		Input:  bridge,
		Source: source,
		Dest:   prefix,
	})
}

func (v *Switch) enableMasq(bridge, source, prefix string) {
	if source == prefix {
		return
	}
	// enable masquerade between source and prefix.
	v.firewall.AddRule(libol.IptRule{
		Table:  NatT,
		Chain:  OlPostC,
		Source: source,
		Dest:   prefix,
		Jump:   MasqueradeC,
	})
	v.firewall.AddRule(libol.IptRule{
		Table:  NatT,
		Chain:  OlPostC,
		Output: bridge,
		Source: prefix,
		Dest:   source,
		Jump:   MasqueradeC,
	})
}

func (v *Switch) preWorker(w *NetworkWorker) {
	vnpCfg := w.cfg.OpenVPN
	if vnpCfg != nil {
		routes := vnpCfg.Routes
		routes = append(routes, vnpCfg.Subnet)
		if addr := w.GetSubnet(); addr != "" {
			routes = append(routes, addr)
		}
		for _, rt := range w.cfg.Routes {
			addr := rt.Prefix
			if _, inet, err := net.ParseCIDR(addr); err == nil {
				routes = append(routes, inet.String())
			}
		}
		vnpCfg.Routes = routes
	}
}

func (v *Switch) preNetwork() {
	crypt := v.cfg.Crypt
	for _, nCfg := range v.cfg.Network {
		name := nCfg.Name
		w := NewNetworkWorker(nCfg, crypt)
		v.worker[name] = w
		brCfg := nCfg.Bridge
		vpnCfg := nCfg.OpenVPN

		v.preWorker(w)
		source := brCfg.Address
		ifAddr := strings.SplitN(source, "/", 2)[0]
		// Enable MASQUERADE for OpenVPN
		if vpnCfg != nil {
			for _, rt := range vpnCfg.Routes {
				v.allowForward(vpnCfg.Device, vpnCfg.Subnet, rt)
				v.enableMasq(vpnCfg.Device, vpnCfg.Subnet, rt)
			}
		}
		if ifAddr == "" {
			continue
		}
		// Enable MASQUERADE, and allowed forward.
		for _, rt := range nCfg.Routes {
			if vpnCfg != nil {
				v.allowForward(brCfg.Name, vpnCfg.Subnet, rt.Prefix)
				v.enableMasq(brCfg.Name, vpnCfg.Subnet, rt.Prefix)
			}
			if rt.NextHop != ifAddr {
				continue
			}
			v.allowForward(brCfg.Name, source, rt.Prefix)
			if rt.Mode == "snat" {
				v.enableMasq(brCfg.Name, source, rt.Prefix)
			}
		}
	}
}

func (v *Switch) preApplication() {
	// Append accessed auth for point
	v.apps.Auth = app.NewAccess(v)
	v.hooks = append(v.hooks, v.apps.Auth.OnFrame)
	// Append request process
	v.apps.Request = app.NewRequest(v)
	v.hooks = append(v.hooks, v.apps.Request.OnFrame)

	inspect := ""
	for _, v := range v.cfg.Inspect {
		inspect += v
	}
	// Check whether inspect neighbor
	if strings.Contains(inspect, "neighbor") {
		v.apps.Neighbor = app.NewNeighbors(v)
		v.hooks = append(v.hooks, v.apps.Neighbor.OnFrame)
	}
	// Check whether inspect online flow by five-tuple.
	if strings.Contains(inspect, "online") {
		v.apps.OnLines = app.NewOnline(v)
		v.hooks = append(v.hooks, v.apps.OnLines.OnFrame)
	}
	for i, h := range v.hooks {
		v.out.Debug("Switch.preApplication: id %d, func %s", i, libol.FunName(h))
	}
}

func (v *Switch) preController() {
	ctrls.Load(v.cfg.ConfDir + "/ctrl.json")
	if ctrls.Ctrl.Name == "" {
		ctrls.Ctrl.Name = v.cfg.Alias
	}
	ctrls.Ctrl.Switcher = v
}

func (v *Switch) LoadPass(file string) {
	reader, err := os.Open(file)
	if err != nil {
		if !os.IsNotExist(err) {
			libol.Warn("Switch.LoadPass %v", err)
		}
		return
	}
	defer reader.Close()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.SplitN(line, ":", 4)
		if len(columns) < 2 {
			continue
		}
		user := columns[0]
		pass := columns[1]
		role := "guest"
		if len(columns) > 2 {
			role = columns[2]
		}
		userObj := &models.User{
			Name:     user,
			Password: pass,
			Role:     role,
		}
		userObj.Update()
		store.User.Add(userObj)
	}
	if err := scanner.Err(); err != nil {
		v.out.Warn("Switch.LoadPass %v", err)
	}
}

func (v *Switch) Initialize() {
	v.lock.Lock()
	defer v.lock.Unlock()

	store.User.SetFile(v.cfg.Password)
	v.preApplication()
	if v.cfg.Http != nil {
		v.http = NewHttp(v)
	}
	v.preNetwork()
	// Controller
	v.preController()
	// FireWall
	v.firewall.Initialize()
	for _, w := range v.worker {
		w.Initialize()
	}
	// Load password for guest access
	if v.cfg.Password != "" {
		v.LoadPass(v.cfg.Password)
	}
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
	v.out.Info("Switch.onClient: %s", client.String())
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
	addr := client.RemoteAddr()
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
	if private == nil {
		return libol.NewErr("point %s notFound.", addr)
	}
	point, ok := private.(*models.Point)
	if !ok {
		return libol.NewErr("point %s notRight.", addr)
	}
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

func (v *Switch) OnClose(client libol.SocketClient) error {
	addr := client.RemoteAddr()
	v.out.Info("Switch.OnClose: %s", addr)
	// already not need support free list for device.
	uuid := store.Point.GetUUID(addr)
	if store.Point.GetAddr(uuid) == addr { // not has newer
		store.Network.DelLease(uuid)
	}
	store.Point.Del(addr)
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
}

func (v *Switch) Stop() {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.out.Debug("Switch.Stop")
	ctrls.Ctrl.Stop()
	// firstly, notify leave to point.
	for p := range store.Point.List() {
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

func (v *Switch) GetBridge(tenant string) (network.Bridger, error) {
	w, ok := v.worker[tenant]
	if !ok {
		return nil, libol.NewErr("bridge %s notFound", tenant)
	}
	return w.bridge, nil
}

func (v *Switch) NewTap(tenant string) (network.Taper, error) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.out.Debug("Switch.NewTap")

	// already not need support free list for device.
	// dropped firstly packages during 15s because of forwarding delay.
	br, err := v.GetBridge(tenant)
	if err != nil {
		v.out.Error("Switch.NewTap: %s", err)
		return nil, err
	}
	dev, err := network.NewTaper(tenant, network.TapConfig{
		Provider: br.Type(),
		Type:     network.TAP,
		VirtBuf:  v.cfg.Queue.VirWrt,
		KernBuf:  v.cfg.Queue.VirSnd,
	})
	if err != nil {
		v.out.Error("Switch.NewTap: %s", err)
		return nil, err
	}
	dev.Up()
	// add new tap to bridge.
	_ = br.AddSlave(dev.Name())
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
		return libol.NewErr("bridge %s notFound", tenant)
	}
	br := w.bridge
	_ = br.DelSlave(dev.Name())
	v.out.Info("Switch.FreeTap: %s", name)
	return nil
}

func (v *Switch) UUID() string {
	if v.uuid == "" {
		v.uuid = libol.GenRandom(13)
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
	name := device.Name()
	v.out.Info("Switch.ReadTap: %s", name)
	done := make(chan bool, 2)
	queue := make(chan *libol.FrameMessage, v.cfg.Queue.TapWr)
	libol.Go(func() {
		for {
			frame := libol.NewFrameMessage()
			n, err := device.Read(frame.Frame())
			if err != nil {
				v.out.Error("Switch.ReadTap: %s", err)
				done <- true
				break
			}
			frame.SetSize(n)
			if v.out.Has(libol.LOG) {
				v.out.Log("Switch.ReadTap: %x\n", frame.Frame()[:n])
			}
			queue <- frame
		}
	})
	defer device.Close()
	for {
		select {
		case frame := <-queue:
			if err := readAt(frame); err != nil {
				v.out.Error("Switch.ReadTap: readAt %s %s", name, err)
				return
			}
		case <-done:
			return
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
	return config.Manager.Switch
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
