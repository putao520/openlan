package config

import (
	"flag"
	"github.com/danieldin95/openlan-go/libol"
	"runtime"
)

type Interface struct {
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	Mtu      int    `json:"mtu" yaml:"mtu"`
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Bridge   string `json:"bridge,omitempty" yaml:"bridge,omitempty"`
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`
}

type Point struct {
	Alias    string    `json:"name,omitempty" yaml:"name,omitempty"`
	Network  string    `json:"network,omitempty" yaml:"network,omitempty"`
	Addr     string    `json:"connection" yaml:"connection"`
	Username string    `json:"username,omitempty" yaml:"username,omitempty"`
	Password string    `json:"password,omitempty" yaml:"password,omitempty"`
	Protocol string    `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	If       Interface `json:"interface" yaml:"interface"`
	Log      Log       `json:"log" yaml:"log"`
	Http     *Http     `json:"http,omitempty" yaml:"http,omitempty"`
	Allowed  bool      `json:"-" yaml:"-"`
	SaveFile string    `json:"-" yaml:"-"`
}

var pointDef = Point{
	Alias:    "",
	Addr:     "openlan.net",
	Protocol: "tls", // tcp, tls, ws and wss etc.
	Log: Log{
		File:    "./point.log",
		Verbose: libol.INFO,
	},
	If: Interface{
		Mtu:      1518,
		Provider: "tap",
		Name:     "",
	},
	Http: &Http{
		Listen: "127.0.0.1:10001",
	},
	SaveFile: "./point.json",
	Network:  "",
	Allowed:  true,
}

func NewPoint() (c *Point) {
	c = &Point{
		Http:    &Http{},
		Allowed: true,
	}
	flag.StringVar(&c.Alias, "alias", pointDef.Alias, "Alias for this point")
	flag.StringVar(&c.Network, "network", pointDef.Network, "Network name")
	flag.StringVar(&c.Addr, "connection", pointDef.Addr, "Virtual switch connect to")
	flag.StringVar(&c.Username, "username", pointDef.Username, "Accessed username")
	flag.StringVar(&c.Password, "password", pointDef.Password, "Accessed password")
	flag.StringVar(&c.Protocol, "protocol", pointDef.Protocol, "Connection protocol")
	flag.IntVar(&c.Log.Verbose, "log:level", pointDef.Log.Verbose, "log level")
	flag.StringVar(&c.Log.File, "log:file", pointDef.Log.File, "log saved to file")
	flag.StringVar(&c.If.Name, "if:name", pointDef.If.Name, "Configure interface name")
	flag.StringVar(&c.If.Address, "if:addr", pointDef.If.Address, "Configure interface address")
	flag.StringVar(&c.If.Bridge, "if:br", pointDef.If.Bridge, "Configure bridge name")
	flag.StringVar(&c.If.Provider, "if:provider", pointDef.If.Provider, "Interface provider")
	flag.StringVar(&c.Http.Listen, "http:listen", pointDef.Http.Listen, "Http listen on")
	flag.StringVar(&c.SaveFile, "conf", pointDef.SaveFile, "the configuration file")
	flag.Parse()

	if err := c.Load(); err != nil {
		libol.Error("NewPoint.load %s", err)
	}
	c.Default()
	libol.Init(c.Log.File, c.Log.Verbose)
	return c
}

func (c *Point) Right() {
	if c.Alias == "" {
		c.Alias = GetAlias()
	}
	RightAddr(&c.Addr, 10002)
	if runtime.GOOS == "darwin" {
		c.If.Provider = "tun"
	}
}

func (c *Point) Default() {
	c.Right()

	//reset zero value to default
	if c.Addr == "" {
		c.Addr = pointDef.Addr
	}
	if c.If.Mtu == 0 {
		c.If.Mtu = pointDef.If.Mtu
	}
}

func (c *Point) Load() error {
	return libol.UnmarshalLoad(c, c.SaveFile)
}

func init() {
	pointDef.Right()
}
