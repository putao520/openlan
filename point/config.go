package point

import (
	"flag"
	"fmt"
	"strings"

	"github.com/lightstar-dev/openlan-go/libol"
)

type Config struct {
	Addr     string `json:"VsAddr"`
	Auth     string `json:"VsAuth"`
	Verbose  int    `json:"Verbose"`
	Ifmtu    int    `json:"IfMtu"`
	Ifaddr   string `json:"IfAddr"`
	Brname   string `json:"IfBridge"`
	Iftun    bool   `json:"IfTun"`
	Ifethsrc string `json:"IfEthSrc"`
	Ifethdst string `json:"IfEthDst"`
	LogFile  string `json:"LogFile"`

	saveFile string
	name     string
	password string
}

var Default = Config{
	Addr:     "openlan.net",
	Auth:     "hi:hi@123$",
	Verbose:  libol.INFO,
	Ifmtu:    1518,
	Ifaddr:   "",
	Iftun:    false,
	Brname:   "",
	saveFile: ".point.json",
	name:     "",
	password: "",
	Ifethdst: "2e:4b:f0:b7:6d:ba",
	Ifethsrc: "",
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
	flag.IntVar(&c.Verbose, "verbose", Default.Verbose, "open verbose")
	flag.IntVar(&c.Ifmtu, "if:mtu", Default.Ifmtu, "the interface MTU include ethernet")
	flag.StringVar(&c.Ifaddr, "if:addr", Default.Ifaddr, "the interface address")
	flag.StringVar(&c.Brname, "if:br", Default.Brname, "the bridge name")
	flag.BoolVar(&c.Iftun, "if:tun", Default.Iftun, "using tun device as interface, otherwise tap")
	flag.StringVar(&c.Ifethdst, "if:ethdst", Default.Ifethdst, "ethernet destination for tun device")
	flag.StringVar(&c.Ifethsrc, "if:ethsrc", Default.Ifethsrc, "ethernet source for tun device")
	flag.StringVar(&c.saveFile, "conf", Default.SaveFile(), "The configuration file")

	flag.Parse()
	libol.Init(c.LogFile, c.Verbose)

	c.Load()
	c.Default()
	c.Save(fmt.Sprintf("%s.cur", c.saveFile))
	str, err := libol.Marshal(c, false)
	if err != nil {
		libol.Error("NewConfig.json error: %s", err)
	}
	libol.Info("NewConfig.json: %s", str)

	return
}

func (c *Config) Default() {
	if c.Auth != "" {
		values := strings.Split(c.Auth, ":")
		c.name = values[0]
		if len(values) > 1 {
			c.password = values[1]
		}
	}

	RightAddr(&c.Addr, 10002)

	//reset zero value to default
	if c.Addr == "" {
		c.Addr = Default.Addr
	}
	if c.Auth == "" {
		c.Auth = Default.Auth
	}
	if c.Ifmtu == 0 {
		c.Ifmtu = Default.Ifmtu
	}
	if c.Ifaddr == "" {
		c.Ifaddr = Default.Ifaddr
	}
}

func (c *Config) Name() string {
	return c.name
}

func (c *Config) Password() string {
	return c.password
}

func (c *Config) SaveFile() string {
	return c.saveFile
}

func (c *Config) Save(file string) error {
	if file == "" {
		file = c.saveFile
	}

	return libol.MarshalSave(c, file, true)
}

func (c *Config) Load() error {
	if err := libol.UnmarshalLoad(c, c.saveFile); err != nil {
		return err
	}

	return nil
}
