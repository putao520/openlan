package config

import (
	"flag"
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"path/filepath"
	"runtime"
	"strings"
)

type Bridge struct {
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

type Network struct {
	Alias    string        `json:"-"`
	Name     string        `json:"name"`
	Bridge   Bridge        `json:"bridge"`
	Subnet   IpSubnet      `json:"subnet"`
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
}

type Cert struct {
	Dir     string `json:"dir"`
	CrtFile string `json:"crt"`
	KeyFile string `json:"key"`
}

func (c *Cert) Right() {
	if c.Dir == "" {
		return
	}
	if c.CrtFile == "" {
		c.CrtFile = fmt.Sprintf("%s/crt.pem", c.Dir)
	}
	if c.KeyFile == "" {
		c.KeyFile = fmt.Sprintf("%s/private.key", c.Dir)
	}
}

type FlowRules struct {
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

type Socks struct {
	Listen string   `json:"listen,omitempty"`
	Auth   Password `json:"auth,omitempty"`
}

type HttpProxy struct {
	Listen string   `json:"listen,omitempty"`
	Auth   Password `json:"auth,omitempty"`
	Cert   *Cert    `json:"cert,omitempty"`
}

type TcpProxy struct {
	Listen string `json:"listen,omitempty"`
	Target string `json:"target,omitempty"`
}

type Proxy struct {
	Socks *Socks      `json:"socks,omitempty"`
	Http  *HttpProxy  `json:"http,omitempty"`
	Tcp   []*TcpProxy `json:"tcp,omitempty"`
}

func (p *Proxy) Right() {
	libol.Debug("Proxy.Right Socks %v", p.Socks)
	if p.Socks != nil {
		RightAddr(&p.Socks.Listen, 11080)
	}
	libol.Debug("Proxy.Right Http %v", p.Http)
	if p.Http != nil {
		RightAddr(&p.Http.Listen, 11082)
		if p.Http.Cert != nil {
			p.Http.Cert.Right()
		}
	}
	libol.Debug("Proxy.Right Tcp %v", p.Tcp)
}

var fd = Perf{
	Point:    1024,
	Neighbor: 1024,
	OnLine:   64,
	Link:     1024,
}

type Perf struct {
	Point    int `json:"point"`
	Neighbor int `json:"neighbor"`
	OnLine   int `json:"online"`
	Link     int `json:"link"`
}

func (p *Perf) Right() {
	if p.Point == 0 {
		p.Point = fd.Point
	}
	if p.Neighbor == 0 {
		p.Neighbor = fd.Neighbor
	}
	if p.OnLine == 0 {
		p.OnLine = fd.OnLine
	}
	if p.Link == 0 {
		p.Link = fd.Link
	}
}

type Switch struct {
	Alias     string      `json:"alias"`
	Perf      *Perf       `json:"perf,omitempty"`
	Protocol  string      `json:"protocol"` // tcp, tls, udp, kcp, ws and wss.
	Listen    string      `json:"listen"`
	Timeout   int         `json:"timeout"`
	Http      *Http       `json:"http,omitempty"`
	Log       Log         `json:"log"`
	Cert      *Cert       `json:"cert,omitempty"`
	Crypt     *Crypt      `json:"crypt,omitempty"`
	Proxy     *Proxy      `json:"proxy,omitempty"`
	PProf     string      `json:"pprof"`
	Network   []*Network  `json:"network,omitempty"`
	FireWall  []FlowRules `json:"firewall,omitempty"`
	Inspect   string      `json:"inspect"`
	ConfDir   string      `json:"-" yaml:"-"`
	TokenFile string      `json:"-" yaml:"-"`
	SaveFile  string      `json:"-" yaml:"-"`
}

var sd = Switch{
	Timeout: 5 * 60,
	Log: Log{
		File:    "./openlan-switch.log",
		Verbose: libol.INFO,
	},
	Http: &Http{
		Listen: "0.0.0.0:10000",
	},
	Listen: "0.0.0.0:10002",
	Perf:   &fd,
}

func NewSwitch() (c Switch) {
	if runtime.GOOS == "linux" {
		sd.Log.File = "/var/log/openlan-switch.log"
	}
	flag.StringVar(&c.Log.File, "log:file", sd.Log.File, "Configure log file")
	flag.IntVar(&c.Log.Verbose, "log:level", sd.Log.Verbose, "Configure log level")
	flag.StringVar(&c.ConfDir, "conf:dir", sd.ConfDir, "Configure virtual switch directory")
	flag.StringVar(&c.PProf, "prof", sd.PProf, "Configure file for CPU prof")
	flag.Parse()
	c.SaveFile = fmt.Sprintf("%s/switch.json", c.ConfDir)
	if err := c.Load(); err != nil {
		libol.Error("NewSwitch.load %s", err)
	}
	libol.SetLogger(c.Log.File, c.Log.Verbose)
	c.Default()
	libol.Debug("NewSwitch %v", c)
	return c
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
	libol.Debug("Switch.Right Proxy %v", c.Proxy)
	if c.Proxy != nil {
		c.Proxy.Right()
	}
	if c.Perf != nil {
		c.Perf.Right()
	} else {
		c.Perf = &fd
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
}
