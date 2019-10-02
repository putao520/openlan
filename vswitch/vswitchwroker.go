package vswitch

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net"
    "os"
    "sort"
    "strings"
    "sync"
    "time"

    "github.com/lightstar-dev/openlan-go/libol"
    "github.com/lightstar-dev/openlan-go/point"
    "github.com/milosgajdos83/tenus"
    "github.com/songgao/water"
)

type Point struct {
    Client *libol.TcpClient
    Device *water.Interface
}

func NewPoint(c *libol.TcpClient, d *water.Interface) (this *Point) {
    this = &Point {
        Client: c,
        Device: d,
    }

    return
}

type VSwitchWroker struct {
    //Public variable
    Server *TcpServer
    Auth *PointAuth
    Request *WithRequest
    Neighbor *Neighborer
    Redis *libol.RedisCli
    EnableRedis bool
    Conf *Config
    
    //Private variable
    verbose int
    br tenus.Bridger
    brip net.IP
    brnet *net.IPNet
    
    keys []int
    hooks map[int] func (*libol.TcpClient, *libol.Frame) error
    ifmtu int

    clientsLock sync.RWMutex
    clients map[*libol.TcpClient] *Point
    usersLock sync.RWMutex
    users map[string] *User
    newtime int64
    brname string
    linksLock sync.RWMutex
    links map[string] *point.Point
}

func NewVSwitchWroker(server *TcpServer, c *Config) (this *VSwitchWroker) {
    this = &VSwitchWroker {
        Server: server,
        Neighbor: nil,
        Redis: libol.NewRedisCli(c.Redis.Addr, c.Redis.Auth, c.Redis.Db),
        EnableRedis: c.Redis.Enable,
        Conf: c,
        verbose: c.Verbose,
        br: nil,
        ifmtu: c.Ifmtu,
        hooks: make(map[int] func (*libol.TcpClient, *libol.Frame) error),
        keys: make([]int, 0, 1024),
        clients: make(map[*libol.TcpClient] *Point, 1024),
        users: make(map[string] *User, 1024),
        newtime: time.Now().Unix(),
        brname: c.Brname,
        links: make(map[string] *point.Point),
    }

    if err := this.Redis.Open(); err != nil {
        libol.Error("NewVSwitchWroker: redis.Open %s", err)
    }
    this.Auth = NewPointAuth(this, c)
    this.Request = NewWithRequest(this, c)
    this.Neighbor = NewNeighborer(this, c)
    this.NewBr()
    this.Register()
    this.LoadUsers()
    this.LoadLinks()

    return 
}

func (this *VSwitchWroker) Register() {
    this.setHook(0x10, this.Neighbor.OnFrame)
    this.setHook(0x00, this.Auth.OnFrame)
    this.setHook(0x01, this.Request.OnFrame)
    this.showHook()
}

func (this *VSwitchWroker) LoadUsers() error {
    file, err := os.Open(this.Conf.Password)
    if err != nil {
        return err
    }

    defer file.Close()
    reader := bufio.NewReader(file)
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            break
        }
        
        values := strings.Split(line, ":")
        if len(values) == 2 {
            user := NewUser(values[0], strings.TrimSpace(values[1]))
            this.AddUser(user)
        }
    }

    return nil
}

func (this *VSwitchWroker) LoadLinks() {
    if this.Conf.Links != nil {
        for _, lc := range this.Conf.Links {
            lc.Default()
            libol.Info("VSwitchWroker.LoadLinks %s", lc)
            this.AddLink(lc)
        }
    }
}

func (this *VSwitchWroker) BrName() string {
    if this.brname == "" {
        addrs := strings.Split(this.Server.Addr, ":")
        if len(addrs) != 2 {
            this.brname = "brol-default"
        } else {
            this.brname = fmt.Sprintf("brol-%s", addrs[1])
        }
    }

    return this.brname
}

func (this *VSwitchWroker) NewBr() {
    var err error
    var br tenus.Bridger

    addr := this.Conf.Ifaddr
    brname := this.BrName()
    br, err = tenus.BridgeFromName(brname)
    if err != nil {
        br, err = tenus.NewBridgeWithName(brname)
        if err != nil {
            libol.Error("VSwitchWroker.newBr: %s", err)
        }
    }

    if err := libol.BrCtlStp(brname, true); err != nil {
        libol.Error("VSwitchWroker.newBr.ctlstp: %s", err)
    }
    
    if err = br.SetLinkUp(); err != nil {
        libol.Error("VSwitchWroker.newBr: %s", err)
    }

    libol.Info("VSwitchWroker.newBr %s", brname)

    if addr != "" {
        ip, net, err := net.ParseCIDR(addr)
        if err != nil {
            libol.Error("VSwitchWroker.newBr.ParseCIDR %s : %s", addr, err)
        }
        if err := br.SetLinkIp(ip, net); err != nil {
            libol.Error("VSwitchWroker.newBr.SetLinkIp %s : %s", brname, err)
        }

        this.brip = ip
        this.brnet = net
    }

    this.br = br
}

func (this *VSwitchWroker) NewTap() (*water.Interface, error) {
    libol.Debug("VSwitchWroker.newTap")  
    ifce, err := water.New(water.Config {
        DeviceType: water.TAP,
    })
    if err != nil {
        libol.Error("VSwitchWroker.newTap: %s", err)
        return nil, err
    }
    
    link, err := tenus.NewLinkFrom(ifce.Name())
    if err != nil {
        libol.Error("VSwitchWroker.newTap: Get ifce %s: %s", ifce.Name(), err)
        return nil, err
    }
    
    if err := link.SetLinkUp(); err != nil {
        libol.Error("VSwitchWroker.newTap: ", err)
    }

    if err := this.br.AddSlaveIfc(link.NetInterface()); err != nil {
        libol.Error("VSwitchWroker.newTap: Switch ifce %s: %s", ifce.Name(), err)
        return nil, err
    }

    libol.Info("VSwitchWroker.newTap %s", ifce.Name())  

    return ifce, nil
}

func (this *VSwitchWroker) Start() {
    go this.Server.GoAccept()
    go this.Server.GoLoop(this.onClient, this.onRecv, this.onClose)
}

func (this *VSwitchWroker) showHook() {
    for _, k := range this.keys {
        libol.Debug("VSwitchWroker.showHool k:%d func: %p", k, this.hooks[k])
    }
} 

func (this *VSwitchWroker) setHook(index int, hook func (*libol.TcpClient, *libol.Frame) error) {
    this.hooks[index] = hook
    this.keys = append(this.keys, index)
    sort.Ints(this.keys)
}

func (this *VSwitchWroker) onHook(client *libol.TcpClient, data []byte) error {
    frame := libol.NewFrame(data)

    for _, k := range this.keys {
        libol.Debug("VSwitchWroker.onHook k:%d", k)
        if f, ok := this.hooks[k]; ok {
            if err := f(client, frame); err != nil {
                return err
            }
        }   
    }

    return nil
}

func (this *VSwitchWroker) handleReq(client *libol.TcpClient, frame *libol.Frame) error {
    return nil
}

func (this *VSwitchWroker) onClient(client *libol.TcpClient) error {
    client.Status = libol.CL_CONNECTED

    libol.Info("VSwitchWroker.onClient: %s", client.Addr) 

    return nil
}

func (this *VSwitchWroker) onRecv(client *libol.TcpClient, data []byte) error {
    //TODO Hook packets such as ARP Learning.
    libol.Debug("VSwitchWroker.onRecv: %s % x", client.Addr, data)

    if err := this.onHook(client, data); err != nil {
        libol.Debug("VSwitchWroker.onRecv: %s dropping by %s", client.Addr, err)
        return err
    }

    point := this.GetPoint(client)
    if point == nil {
        return libol.Errer("Point not found.")
    }

    ifce := point.Device
    if point == nil || point.Device == nil {
        return libol.Errer("Tap devices is nil")
    }
 
    if _, err := ifce.Write(data); err != nil {
        libol.Error("VSwitchWroker.onRecv: %s", err)
        return err
    }

    return nil
}

func (this *VSwitchWroker) onClose(client *libol.TcpClient) error {
    libol.Info("VSwitchWroker.onClose: %s", client.Addr)

    this.DelPoint(client)
    
    return nil
}

func (this *VSwitchWroker) Close() {
    libol.Info("VSwitchWroker.Close")

    this.Server.Close()

    if this.br != nil && this.brip != nil {
        if err := this.br.UnsetLinkIp(this.brip, this.brnet); err != nil {
            libol.Error("VSwitchWroker.Close.UnsetLinkIp %s : %s", this.br.NetInterface().Name, err)
        }
    }

    for _, p := range this.links {
        p.Close()
    }
}

func (this *VSwitchWroker) AddUser(user *User) {
    this.usersLock.Lock()
    defer this.usersLock.Unlock()

    name := user.Name 
    if name == "" {
        name = user.Token
    }
    this.users[name] = user
}

func (this *VSwitchWroker) GetUser(name string) *User {
    this.usersLock.RLock()
    defer this.usersLock.RUnlock()

    if u, ok := this.users[name]; ok {
        return u
    }

    return nil
}

func (this *VSwitchWroker) ListUser() chan *User {
    c := make(chan *User, 128)

    go func() {
        this.usersLock.RLock()
        defer this.usersLock.RUnlock()

        for _, u := range this.users {
            c <- u
        }
        c <- nil //Finish channel by nil.
    }()

    return c
}

func (this *VSwitchWroker) AddPoint(p *Point) {
    this.clientsLock.Lock()
    defer this.clientsLock.Unlock()

    this.PubPoint(p, true)
    this.clients[p.Client] = p
}

func (this *VSwitchWroker) GetPoint(c *libol.TcpClient) *Point {
    this.clientsLock.RLock()
    defer this.clientsLock.RUnlock()

    if p, ok := this.clients[c]; ok {
        return p
    }
    return nil
}

func (this *VSwitchWroker) DelPoint(c *libol.TcpClient) {
    this.clientsLock.Lock()
    defer this.clientsLock.Unlock()
    
    if p, ok := this.clients[c]; ok {
        p.Device.Close()
        this.PubPoint(p, false)
        delete(this.clients, c)
    }
}

func (this *VSwitchWroker) ListPoint() chan *Point {
    c := make(chan *Point, 128)

    go func() {
        this.clientsLock.RLock()
        defer this.clientsLock.RUnlock()

        for _, p := range this.clients {
            c <- p
        }
        c <- nil //Finish channel by nil.
    }()

    return c
}

func (this *VSwitchWroker) UpTime() int64 {
    return time.Now().Unix() - this.newtime
}

func (this *VSwitchWroker) PubPoint(p *Point, isadd bool) {
    if !this.EnableRedis {
        return
    }

    key := fmt.Sprintf("point:%s", strings.Replace(p.Client.String(), ":", "-", -1))
    value := map[string] interface{} {
        "remote": p.Client.String(), 
        "newtime": p.Client.NewTime,
        "device": p.Device.Name(),
        "actived": isadd, 
    }

    if err := this.Redis.HMSet(key, value); err != nil {
        libol.Error("VSwitchWroker.PubPoint hset %s", err)
    }
}

func (this *VSwitchWroker) AddLink(c *point.Config) {
    c.Brname = this.BrName() //Reset bridge name.

    go func() {
        p := point.NewPoint(c)

        this.linksLock.Lock()
        this.links[c.Addr] = p
        this.linksLock.Unlock()

        p.UpLink()
        p.Start()
    }()
}

func (this *VSwitchWroker) DelLink(addr string) {
    this.linksLock.Lock()
    defer this.linksLock.Unlock()
    if p, ok := this.links[addr]; ok {
        p.Close()
        delete(this.links, addr)
    }
}

func (this *VSwitchWroker) GetLink(addr string) *point.Point {
    this.linksLock.RLock()
    defer this.linksLock.RUnlock()

    if p, ok := this.links[addr]; ok {
        return p
    }
    
    return nil
}

func (this *VSwitchWroker) ListLink() chan *point.Point {
    c := make(chan *point.Point, 128)

    go func() {
        this.linksLock.RLock()
        defer this.linksLock.RUnlock()

        for _, p := range this.links {
            c <- p
        }
        c <- nil //Finish channel by nil.
    }()

    return c
}

type PointAuth struct {
    ifmtu int
    wroker *VSwitchWroker
}

func NewPointAuth(wroker *VSwitchWroker, c *Config) (this *PointAuth) {
    this = &PointAuth {
        ifmtu: c.Ifmtu,
        wroker: wroker,
    }
    return
}

func (this *PointAuth) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
    libol.Debug("PointAuth.OnFrame % x.", frame.Data)

    if libol.IsInst(frame.Data) {
        action := libol.DecAction(frame.Data)
        libol.Debug("PointAuth.OnFrame.action: %s", action)

        if action == "logi=" {
            if err := this.handleLogin(client, libol.DecBody(frame.Data)); err != nil {
                libol.Error("PointAuth.OnFrame: %s", err)
                client.SendResp("login", err.Error())
                client.Close()
                return err
            }
            client.SendResp("login", "okay.")
        }

        return nil
    }

    if client.Status != libol.CL_AUTHED {
        client.Droped++
        libol.Debug("PointAuth.onRecv: %s unauth", client.Addr)
        return libol.Errer("Unauthed client.")
    }

    return nil
}

func  (this *PointAuth) handleLogin(client *libol.TcpClient, data string) error {
    libol.Debug("PointAuth.handleLogin: %s", data)

    if client.Status == libol.CL_AUTHED {
        libol.Warn("PointAuth.handleLogin: already authed %s", client)
        return nil
    }
    
    user := NewUser("", "")
    if err := json.Unmarshal([]byte(data), user); err != nil {
        return libol.Errer("Invalid json data.")
    }

    name := user.Name
    if user.Token != "" {
        name = user.Token
    }
    _user := this.wroker.GetUser(name)
    if _user != nil {
        if _user.Password == user.Password {
            client.Status = libol.CL_AUTHED
            libol.Info("PointAuth.handleLogin: %s Authed", client.Addr)
            this.onAuth(client)
            return nil
        }

        client.Status = libol.CL_UNAUTH
    }

    return libol.Errer("Auth failed.")
}

func (this *PointAuth) onAuth(client *libol.TcpClient) error {
    if client.Status != libol.CL_AUTHED {
        return libol.Errer("not authed.")
    }

    libol.Info("PointAuth.onAuth: %s", client.Addr)
    ifce, err := this.wroker.NewTap()
    if err != nil {
        return err
    }

    this.wroker.AddPoint(NewPoint(client, ifce))
    
    go this.GoRecv(ifce, client.SendMsg)

    return nil
}

func (this *PointAuth) GoRecv(ifce *water.Interface, dorecv func ([]byte) error) {
    libol.Info("PointAuth.GoRecv: %s", ifce.Name())

    defer ifce.Close()
    for {
        data := make([]byte, this.ifmtu)
        n, err := ifce.Read(data)
        if err != nil {
            libol.Error("PointAuth.GoRev: %s", err)
            break
        }

        libol.Debug("PointAuth.GoRev: % x\n", data[:n])
        if err := dorecv(data[:n]); err != nil {
            libol.Error("PointAuth.GoRev: do-recv %s %s", ifce.Name(), err)
        }
    }
}

type WithRequest struct {
    wroker *VSwitchWroker
}

func NewWithRequest(wroker *VSwitchWroker, c *Config) (this *WithRequest) {
    this = &WithRequest {
        wroker: wroker,
    }
    return
}

func (this *WithRequest) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
    libol.Debug("WithRequest.OnFrame % x.", frame.Data)

    if libol.IsInst(frame.Data) {
        action, body := libol.DecActionBody(frame.Data)
        libol.Debug("WithRequest.OnFrame.action: %s %s", action, body)

        if action == "neig=" {
            //TODO
        }
    }

    return nil
}

