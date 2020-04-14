package config

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/danieldin95/openlan-go/libol"
)

type Point struct {
	Alias    string `json:"alias"`
	Addr     string `json:"vs.addr"`
	Auth     string `json:"vs.auth"`
	Tls      bool   `json:"vs.tls"`
	Allowed  bool   `json:"vs.allowed"`
	IfName   string `json:"if.name"`
	IfMtu    int    `json:"if.mtu"`
	IfAddr   string `json:"if.addr"`
	BrName   string `json:"if.br"`
	IfTun    bool   `json:"if.tun"`
	LogFile  string `json:"log.file"`
	Verbose  int    `json:"log.level"`
	Script   string `json:"script"`
	Network  string `json:"network"`

	SaveFile string `json:"-"`
	name     string
	password string
}

var PointDefault = Point{
	Alias:    "",
	Addr:     "openlan.net",
	Auth:     "hi:123456",
	Verbose:  libol.INFO,
	IfMtu:    1518,
	IfAddr:   "",
	IfTun:    false,
	BrName:   "",
	SaveFile: "point.json",
	Network:  "",
	name:     "",
	password: "",
	LogFile:  "point.log",
	Script:   fmt.Sprintf("point.%s.cmd", runtime.GOOS),
	Allowed:  true,
}

func NewPoint() (c *Point) {
	c = &Point{
		LogFile: PointDefault.LogFile,
	}

	flag.StringVar(&c.Alias, "alias", PointDefault.Alias, "the alias for this point")
	flag.StringVar(&c.Network, "network", PointDefault.Network, "the network login to")
	flag.StringVar(&c.Addr, "vs:addr", PointDefault.Addr, "the server connect to")
	flag.StringVar(&c.Auth, "vs:auth", PointDefault.Auth, "the auth login to")
	flag.BoolVar(&c.Tls, "vs:tls", PointDefault.Tls, "enable TLS to decrypt")
	flag.IntVar(&c.Verbose, "log:level", PointDefault.Verbose, "logger level")
	flag.StringVar(&c.LogFile, "log:file", PointDefault.LogFile, "logger file")
	flag.StringVar(&c.IfName, "if:name", PointDefault.IfAddr, "the interface name when has multiple device.")
	flag.StringVar(&c.IfAddr, "if:addr", PointDefault.IfAddr, "the interface address")
	flag.StringVar(&c.BrName, "if:br", PointDefault.BrName, "the bridge name")
	flag.BoolVar(&c.IfTun, "if:tun", PointDefault.IfTun, "using tun device as interface, otherwise tap")
	flag.StringVar(&c.SaveFile, "conf", PointDefault.SaveFile, "the configuration file")
	flag.StringVar(&c.Script, "script", PointDefault.Script, "call script you assigned")
	flag.BoolVar(&c.Allowed, "vs:allowed", PointDefault.Allowed, "allowed network configuration from vs")
	flag.Parse()

	if err := c.Load(); err != nil {
		libol.Error("NewPoint.load %s", err)
	}
	c.Default()
	libol.Init(c.LogFile, c.Verbose)
	_ = c.Save(fmt.Sprintf("%s.cur", c.SaveFile))

	str, err := libol.Marshal(c, false)
	if err != nil {
		libol.Error("NewPoint.json error: %s", err)
	}
	libol.Debug("NewPoint.json: %s", str)

	return
}

func (c *Point) Right() {
	if c.Alias == "" {
		c.Alias = GetAlias()
	}
	if c.Auth != "" {
		values := strings.Split(c.Auth, ":")
		c.name = values[0]
		if len(values) > 1 {
			c.password = values[1]
		}
		values = strings.Split(c.name, "@")
		if len(values) > 1 {
			c.name = values[0]
			c.Network = values[1]
		}
	}
	RightAddr(&c.Addr, 10002)
	if runtime.GOOS == "darwin" {
		c.IfTun = true
	}
}

func (c *Point) Default() {
	c.Right()

	//reset zero value to default
	if c.Addr == "" {
		c.Addr = PointDefault.Addr
	}
	if c.Auth == "" {
		c.Auth = PointDefault.Auth
	}
	if c.IfMtu == 0 {
		c.IfMtu = PointDefault.IfMtu
	}
	if c.IfAddr == "" {
		c.IfAddr = PointDefault.IfAddr
	}
	if c.Network == "" {
		c.Network = PointDefault.Network
	}
}

func (c *Point) Name() string {
	return c.name
}

func (c *Point) Password() string {
	return c.password
}

func (c *Point) Save(file string) error {
	if file == "" {
		file = c.SaveFile
	}

	return libol.MarshalSave(c, file, true)
}

func (c *Point) Load() error {
	if err := libol.UnmarshalLoad(c, c.SaveFile); err != nil {
		return err
	}

	return nil
}

func init() {
	PointDefault.Right()
}
