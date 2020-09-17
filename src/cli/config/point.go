package config

import (
	"flag"
	"github.com/danieldin95/openlan-go/src/libol"
	"runtime"
)

type Interface struct {
	Name     string `json:"name,omitempty"`
	IfMtu    int    `json:"mtu"`
	Address  string `json:"address,omitempty"`
	Bridge   string `json:"bridge,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type Point struct {
	Alias       string    `json:"alias,omitempty"`
	Network     string    `json:"network,omitempty"`
	Connection  string    `json:"connection"`
	Timeout     int       `json:"timeout"`
	Username    string    `json:"username,omitempty"`
	Password    string    `json:"password,omitempty"`
	Protocol    string    `json:"protocol,omitempty"`
	Interface   Interface `json:"interface"`
	Log         Log       `json:"log"`
	Http        *Http     `json:"http,omitempty"`
	Crypt       *Crypt    `json:"crypt,omitempty"`
	PProf       string    `json:"pprof"`
	RequestAddr bool      `json:"-" yaml:"-"`
	SaveFile    string    `json:"-" yaml:"-"`
	Queue       *Queue    `json:"queue"`
	Terminal    string    `json:"-"`
}

var pd = &Point{
	Alias:      "",
	Connection: "openlan.net",
	Protocol:   "tls", // udp, kcp, tcp, tls, ws and wss etc.
	Timeout:    60,
	Log: Log{
		File:    "./openlan-point.log",
		Verbose: libol.INFO,
	},
	Interface: Interface{
		IfMtu:    1518,
		Provider: "tap",
		Name:     "",
	},
	SaveFile:    "./point.json",
	Network:     "default",
	RequestAddr: true,
	Crypt:       &Crypt{},
	Terminal:    "on",
}

func NewPoint() (c *Point) {
	c = &Point{
		RequestAddr: true,
		Crypt:       &Crypt{},
	}
	flag.StringVar(&c.Alias, "alias", pd.Alias, "alias for this point")
	flag.StringVar(&c.Terminal, "terminal", pd.Terminal, "run interactive terminal")
	flag.StringVar(&c.Network, "net", pd.Network, "Network name")
	flag.StringVar(&c.Connection, "conn", pd.Connection, "Virtual switch connect to")
	flag.StringVar(&c.Username, "user", pd.Username, "Accessed username")
	flag.StringVar(&c.Password, "pass", pd.Password, "Accessed password")
	flag.StringVar(&c.Protocol, "proto", pd.Protocol, "Connection protocol")
	flag.IntVar(&c.Timeout, "timeout", pd.Timeout, "Time in secs socket dead")
	flag.IntVar(&c.Log.Verbose, "log:level", pd.Log.Verbose, "Log level")
	flag.StringVar(&c.Log.File, "log:file", pd.Log.File, "Log saved to file")
	flag.StringVar(&c.Interface.Name, "if:name", pd.Interface.Name, "Configure interface name")
	flag.StringVar(&c.Interface.Address, "if:addr", pd.Interface.Address, "Configure interface address")
	flag.StringVar(&c.Interface.Bridge, "if:br", pd.Interface.Bridge, "Configure bridge name")
	flag.StringVar(&c.Interface.Provider, "if:provider", pd.Interface.Provider, "Interface provider")
	flag.StringVar(&c.SaveFile, "conf", pd.SaveFile, "the configuration file")
	flag.StringVar(&c.Crypt.Secret, "crypt:secret", pd.Crypt.Secret, "Crypt secret")
	flag.StringVar(&c.Crypt.Algo, "crypt:algo", pd.Crypt.Algo, "Crypt algorithm")
	flag.StringVar(&c.PProf, "pprof", pd.PProf, "Configure file for CPU prof")
	flag.Parse()
	c.Initialize()
	return c
}

func (c *Point) Id() string {
	return c.Connection + ":" + c.Network
}

func (c *Point) Initialize() {
	if err := c.Load(); err != nil {
		libol.Warn("NewPoint.load %s", err)
	}
	c.Default()
	libol.SetLogger(c.Log.File, c.Log.Verbose)
}

func (c *Point) Right() {
	if c.Alias == "" {
		c.Alias = GetAlias()
	}
	RightAddr(&c.Connection, 10002)
	if runtime.GOOS == "darwin" {
		c.Interface.Provider = "tun"
	}
}

func (c *Point) Default() {
	c.Right()
	if c.Queue == nil {
		c.Queue = &Queue{}
	}
	c.Queue.Default()
	//reset zero value to default
	if c.Connection == "" {
		c.Connection = pd.Connection
	}
	if c.Interface.IfMtu == 0 {
		c.Interface.IfMtu = pd.Interface.IfMtu
	}
	if c.Timeout == 0 {
		c.Timeout = pd.Timeout
	}
	if c.Crypt != nil {
		c.Crypt.Default()
	}
}

func (c *Point) Load() error {
	if err := libol.FileExist(c.SaveFile); err == nil {
		return libol.UnmarshalLoad(c, c.SaveFile)
	}
	return nil
}

func init() {
	pd.Right()
	if runtime.GOOS == "linux" {
		pd.Log.File = "/var/log/openlan-point.log"
	}
}
