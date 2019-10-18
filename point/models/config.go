package models

import (
	"flag"
	"fmt"
	"strings"

	"github.com/lightstar-dev/openlan-go/libol"
)

type Config struct {
	Addr     string `json:"VsAddr,omitempty"`
	Auth     string `json:"VsAuth,omitempty"`
	Tls      bool   `json:"VsTls,omitempty"`
	Verbose  int    `json:"Verbose,omitempty"`
	IfMtu    int    `json:"IfMtu,omitempty"`
	IfAddr   string `json:"IfAddr,omitempty"`
	BrName   string `json:"IfBridge,omitempty"`
	IfTun    bool   `json:"IfTun,omitempty"`
	IfEthSrc string `json:"IfEthSrc,omitempty"`
	IfEthDst string `json:"IfEthDst,omitempty"`
	LogFile  string `json:"LogFile,omitempty"`

	SaveFile string `json:"-"`
	name     string
	password string
}

var Default = Config{
	Addr:     "openlan.net",
	Auth:     "hi:hi@123$",
	Verbose:  libol.INFO,
	IfMtu:    1518,
	IfAddr:   "",
	IfTun:    false,
	BrName:   "",
	SaveFile: ".point.json",
	name:     "",
	password: "",
	IfEthDst: "2e:4b:f0:b7:6d:ba",
	IfEthSrc: "",
	LogFile:  ".point.error",
}

func RightAddr(listen *string, port int) {
	values := strings.Split(*listen, ":")
	if len(values) == 1 {
		*listen = fmt.Sprintf("%s:%d", values[0], port)
	}
}

func NewConfig() (c *Config) {
	c = &Config{
		LogFile: Default.LogFile,
	}

	flag.StringVar(&c.Addr, "vs:addr", Default.Addr, "the server connect to")
	flag.StringVar(&c.Auth, "vs:auth", Default.Auth, "the auth login to")
	flag.BoolVar(&c.Tls, "vs:tls", Default.Tls, "Enable TLS to decrypt")
	flag.IntVar(&c.Verbose, "verbose", Default.Verbose, "open verbose")
	flag.IntVar(&c.IfMtu, "if:mtu", Default.IfMtu, "the interface MTU include ethernet")
	flag.StringVar(&c.IfAddr, "if:addr", Default.IfAddr, "the interface address")
	flag.StringVar(&c.BrName, "if:br", Default.BrName, "the bridge name")
	flag.BoolVar(&c.IfTun, "if:tun", Default.IfTun, "using tun device as interface, otherwise tap")
	flag.StringVar(&c.IfEthDst, "if:ethdst", Default.IfEthDst, "ethernet destination for tun device")
	flag.StringVar(&c.IfEthSrc, "if:ethsrc", Default.IfEthSrc, "ethernet source for tun device")
	flag.StringVar(&c.SaveFile, "conf", Default.SaveFile, "The configuration file")

	flag.Parse()
	if err := c.Load(); err != nil {
		libol.Error("NewConfig.load %s", err)
	}
	c.Default()

	libol.Init(c.LogFile, c.Verbose)
	c.Save(fmt.Sprintf("%s.cur", c.SaveFile))

	str, err := libol.Marshal(c, false)
	if err != nil {
		libol.Error("NewConfig.json error: %s", err)
	}
	libol.Debug("NewConfig.json: %s", str)

	return
}

func (c *Config) Right() {
	if c.Auth != "" {
		values := strings.Split(c.Auth, ":")
		c.name = values[0]
		if len(values) > 1 {
			c.password = values[1]
		}
	}
	RightAddr(&c.Addr, 10002)
}

func (c *Config) Default() {
	c.Right()

	//reset zero value to default
	if c.Addr == "" {
		c.Addr = Default.Addr
	}
	if c.Auth == "" {
		c.Auth = Default.Auth
	}
	if c.IfMtu == 0 {
		c.IfMtu = Default.IfMtu
	}
	if c.IfAddr == "" {
		c.IfAddr = Default.IfAddr
	}
}

func (c *Config) Name() string {
	return c.name
}

func (c *Config) Password() string {
	return c.password
}

func (c *Config) Save(file string) error {
	if file == "" {
		file = c.SaveFile
	}

	return libol.MarshalSave(c, file, true)
}

func (c *Config) Load() error {
	if err := libol.UnmarshalLoad(c, c.SaveFile); err != nil {
		return err
	}

	return nil
}

func init() {
	Default.Right()
}
