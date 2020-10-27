package config

import (
	"flag"
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Bridge struct {
	Peer     string `json:"peer"`
	Name     string `json:"name"`
	IfMtu    int    `json:"mtu"`
	Address  string `json:"address,omitempty"`
	Provider string `json:"provider"`
	Stp      string `json:"stp"`
	Delay    int    `json:"delay"`
}

type IpSubnet struct {
	Start   string `json:"start"`
	End     string `json:"end"`
	Netmask string `json:"netmask"`
}

type PrefixRoute struct {
	Prefix  string `json:"prefix"`
	NextHop string `json:"nexthop"`
	Metric  int    `json:"metric"`
}

type HostLease struct {
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
}

type Password struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type OpenVPN struct {
	Name      string   `json:"-"`
	WorkDir   string   `json:"-"`
	Listen    string   `json:"listen"`
	Protocol  string   `json:"protocol"`
	Subnet    string   `json:"subnet"`
	Device    string   `json:"device"`
	Auth      string   `json:"auth"` // xauth or cert.
	DhPem     string   `json:"dh"`
	RootCa    string   `json:"ca"`
	ServerCrt string   `json:"cert"`
	ServerKey string   `json:"key"`
	TlsAuth   string   `json:"tlsauth"`
	Cipher    string   `json:"cipher"`
	Routes    []string `json:"routes"`
	Script    string   `json:"-"`
}

var defaultOvpn = OpenVPN{
	Protocol:  "tcp",
	Auth:      "xauth",
	Device:    "tun",
	RootCa:    "/var/openlan/cert/ca.crt",
	ServerCrt: "/var/openlan/cert/crt",
	ServerKey: "/var/openlan/cert/key",
	DhPem:     "/var/openlan/openvpn/dh.pem",
	TlsAuth:   "/var/openlan/openvpn/ta.key",
	Cipher:    "AES-256-CBC",
	Script:    "/usr/bin/openlan-checkpass " + strings.Join(os.Args[1:], " "),
}

func (o *OpenVPN) Right() {
	if o.WorkDir == "" {
		o.WorkDir = "/var/openlan/openvpn/" + o.Name
	}
	if o.Auth == "" {
		o.Auth = defaultOvpn.Auth
	}
	if o.Device == "" {
		o.Device = defaultOvpn.Device
	}
	if o.Protocol == "" {
		o.Protocol = defaultOvpn.Protocol
	}
	if o.DhPem == "" {
		o.DhPem = defaultOvpn.DhPem
	}
	if o.RootCa == "" {
		o.RootCa = defaultOvpn.RootCa
	}
	if o.ServerCrt == "" {
		o.ServerCrt = defaultOvpn.ServerCrt
	}
	if o.ServerKey == "" {
		o.ServerKey = defaultOvpn.ServerKey
	}
	if o.TlsAuth == "" {
		o.TlsAuth = defaultOvpn.TlsAuth
	}
	if o.Cipher == "" {
		o.Cipher = defaultOvpn.Cipher
	}
	o.Script = defaultOvpn.Script
}

type Network struct {
	Alias    string        `json:"-"`
	Name     string        `json:"name,omitempty"`
	Bridge   Bridge        `json:"bridge,omitempty"`
	Subnet   IpSubnet      `json:"subnet,omitempty"`
	OpenVPN  *OpenVPN      `json:"openvpn,omitempty"`
	Links    []*Point      `json:"links,omitempty"`
	Hosts    []HostLease   `json:"hosts,omitempty"`
	Routes   []PrefixRoute `json:"routes,omitempty"`
	Password []Password    `json:"password,omitempty"`
}

func (n *Network) Right() {
	if n.Bridge.Name == "" {
		n.Bridge.Name = "br-" + n.Name
	}
	if n.Bridge.Provider == "" {
		n.Bridge.Provider = "linux"
	}
	if n.Bridge.IfMtu == 0 {
		n.Bridge.IfMtu = 1518
	}
	if n.Bridge.Delay == 0 {
		n.Bridge.Delay = 2
	}
	if n.Bridge.Stp == "" {
		n.Bridge.Stp = "on"
	}
	ifAddr := strings.SplitN(n.Bridge.Address, "/", 2)[0]
	for i := range n.Routes {
		if n.Routes[i].Metric == 0 {
			n.Routes[i].Metric = 666
		}
		if n.Routes[i].NextHop == "" {
			n.Routes[i].NextHop = ifAddr
		}
	}
	if n.OpenVPN != nil {
		n.OpenVPN.Name = n.Name
		n.OpenVPN.Right()
	}
}

type FlowRule struct {
	Table    string `json:"table"`
	Chain    string `json:"chain"`
	Input    string `json:"input"`
	Source   string `json:"source"`
	ToSource string `json:"to-source"`
	Dest     string `json:"destination"`
	ToDest   string `json:"to-destination"`
	Output   string `json:"output"`
	Comment  string `json:"comment"`
	Jump     string `json:"jump"` // SNAT/RETURN/MASQUERADE
}

var pfd = Perf{
	Point:    1024,
	Neighbor: 1024,
	OnLine:   64,
	Link:     1024,
	User:     1024,
}

type Perf struct {
	Point    int `json:"point"`
	Neighbor int `json:"neighbor"`
	OnLine   int `json:"online"`
	Link     int `json:"link"`
	User     int `json:"user"`
}

func (p *Perf) Right() {
	if p.Point == 0 {
		p.Point = pfd.Point
	}
	if p.Neighbor == 0 {
		p.Neighbor = pfd.Neighbor
	}
	if p.OnLine == 0 {
		p.OnLine = pfd.OnLine
	}
	if p.Link == 0 {
		p.Link = pfd.Link
	}
	if p.User == 0 {
		p.User = pfd.User
	}
}

type Switch struct {
	Alias     string     `json:"alias"`
	Perf      *Perf      `json:"perf,omitempty"`
	Protocol  string     `json:"protocol"` // tcp, tls, udp, kcp, ws and wss.
	Listen    string     `json:"listen"`
	Timeout   int        `json:"timeout"`
	Http      *Http      `json:"http,omitempty"`
	Log       Log        `json:"log"`
	Cert      *Cert      `json:"cert,omitempty"`
	Crypt     *Crypt     `json:"crypt,omitempty"`
	PProf     string     `json:"pprof"`
	Network   []*Network `json:"network,omitempty"`
	FireWall  []FlowRule `json:"firewall,omitempty"`
	Inspect   []string   `json:"inspect"`
	Queue     *Queue     `json:"queue"`
	ConfDir   string     `json:"-"`
	TokenFile string     `json:"-"`
	SaveFile  string     `json:"-"`
}

var sd = &Switch{
	Timeout: 120,
	Log: Log{
		File:    "./openlan-switch.log",
		Verbose: libol.INFO,
	},
	Http: &Http{
		Listen: "0.0.0.0:10000",
	},
	Listen: "0.0.0.0:10002",
	Perf:   &pfd,
}

func NewSwitch() (c Switch) {
	flag.StringVar(&c.Log.File, "log:file", sd.Log.File, "Configure log file")
	flag.StringVar(&c.ConfDir, "conf:dir", sd.ConfDir, "Configure virtual switch directory")
	flag.StringVar(&c.PProf, "prof", sd.PProf, "Http listen for CPU prof")
	flag.IntVar(&c.Log.Verbose, "log:level", sd.Log.Verbose, "Configure log level")
	flag.Parse()
	c.Initialize()
	return c
}

func (c *Switch) Initialize() {
	c.SaveFile = fmt.Sprintf("%s/switch.json", c.ConfDir)
	if err := c.Load(); err != nil {
		libol.Error("Switch.Initialize %s", err)
	}
	c.Default()
	libol.Debug("Switch.Initialize %v", c)
}

func (c *Switch) Right() {
	if c.Alias == "" {
		c.Alias = GetAlias()
	}
	RightAddr(&c.Listen, 10002)
	if c.Http != nil {
		RightAddr(&c.Http.Listen, 10000)
	}
	libol.Debug("Proxy.Right Http %v", c.Http)
	c.TokenFile = fmt.Sprintf("%s/token", c.ConfDir)
	c.SaveFile = fmt.Sprintf("%s/switch.json", c.ConfDir)
	if c.Cert != nil {
		c.Cert.Right()
		// default is tls if cert configured
		if c.Protocol == "" {
			c.Protocol = "tls"
		}
	}
	if c.Perf != nil {
		c.Perf.Right()
	} else {
		c.Perf = &pfd
	}
}

func (c *Switch) Default() {
	c.Right()
	if c.Network == nil {
		c.Network = make([]*Network, 0, 32)
	}
	if c.Timeout == 0 {
		c.Timeout = sd.Timeout
	}
	if c.Crypt != nil {
		c.Crypt.Default()
	}
	if c.Queue == nil {
		c.Queue = &Queue{}
	}
	c.Queue.Default()
	files, err := filepath.Glob(c.ConfDir + "/network/*.json")
	if err != nil {
		libol.Error("Switch.Default %s", err)
	}
	for _, k := range files {
		n := &Network{
			Alias: c.Alias,
		}
		if err := libol.UnmarshalLoad(n, k); err != nil {
			libol.Error("Switch.Default %s", err)
			continue
		}
		c.Network = append(c.Network, n)
	}
	for _, n := range c.Network {
		for _, link := range n.Links {
			link.Default()
		}
		n.Right()
		n.Alias = c.Alias
	}
}

func (c *Switch) Load() error {
	return libol.UnmarshalLoad(c, c.SaveFile)
}

func init() {
	sd.Right()
	if runtime.GOOS == "linux" {
		sd.Log.File = "/var/log/openlan-switch.log"
	}
}
