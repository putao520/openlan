package models

import (
	"flag"
	"fmt"
	"github.com/lightstar-dev/openlan-go/point/models"
	"strings"

	"github.com/lightstar-dev/openlan-go/libol"
)

type Config struct {
	TcpListen  string           `json:"Listen,omitempty"`
	Verbose    int              `json:"Verbose,omitempty"`
	HttpListen string           `json:"Http,omitempty"`
	IfMtu      int              `json:"IfMtu,omitempty"`
	IfAddr     string           `json:"IfAddr,omitempty"`
	BrName     string           `json:"IfBridge,omitempty"`
	Token      string           `json:"AdminToken,omitempty"`
	TokenFile  string           `json:"AdminFile,omitempty"`
	Password   string           `json:"AuthFile,omitempty"`
	Redis      RedisConfig      `json:"Redis,omitempty"`
	LogFile    string           `json:"LogFile,omitempty"`
	CrtFile    string           `json:"CrtFile,omitempty"`
	KeyFile    string           `json:"KeyFile,omitempty"`
	Links      []*models.Config `json:"Links,omitempty"`
	SaveFile   string           `json:"-"`
}

type RedisConfig struct {
	Enable bool   `json:"Enable,omitempty"`
	Addr   string `json:"Addr,omitempty"`
	Auth   string `json:"Auth,omitempty"`
	Db     int    `json:"Database,omitempty"`
}

var Default = Config{
	BrName:     "",
	Verbose:    libol.INFO,
	HttpListen: "",
	TcpListen:  "0.0.0.0:10002",
	Token:      "",
	TokenFile:  ".vswitch.token",
	Password:   ".password",
	IfMtu:      1518,
	IfAddr:     "",
	Redis: RedisConfig{
		Addr:   "127.0.0.1",
		Auth:   "",
		Db:     0,
		Enable: false,
	},
	LogFile:  ".vswitch.error",
	SaveFile: ".vswitch.json",
	CrtFile:  "",
	KeyFile:  "",
	Links:    nil,
}

func RightAddr(listen *string, port int) {
	values := strings.Split(*listen, ":")
	if len(values) == 1 {
		*listen = fmt.Sprintf("%s:%d", values[0], port)
	}
}

func NewConfig() (c *Config) {
	c = &Config{
		Redis:   Default.Redis,
		LogFile: Default.LogFile,
	}

	flag.IntVar(&c.Verbose, "verbose", Default.Verbose, "open verbose")
	flag.StringVar(&c.HttpListen, "http:addr", Default.HttpListen, "the http listen on")
	flag.StringVar(&c.TcpListen, "vs:addr", Default.TcpListen, "the server listen on")
	flag.StringVar(&c.Token, "admin:token", Default.Token, "Administrator token")
	flag.StringVar(&c.TokenFile, "admin:file", Default.TokenFile, "The file administrator token saved to")
	flag.StringVar(&c.Password, "auth:file", Default.Password, "The file password loading from.")
	flag.IntVar(&c.IfMtu, "if:mtu", Default.IfMtu, "the interface MTU include ethernet")
	flag.StringVar(&c.IfAddr, "if:addr", Default.IfAddr, "the interface address")
	flag.StringVar(&c.BrName, "if:br", Default.BrName, "the bridge name")
	flag.StringVar(&c.SaveFile, "conf", Default.SaveFile, "The configuration file")
	flag.StringVar(&c.CrtFile, "tls:crt", Default.CrtFile, "The X509 certificate file for TLS")
	flag.StringVar(&c.KeyFile, "tls:key", Default.KeyFile, "The X509 certificate key for TLS")

	flag.Parse()
	c.Default()
	if err := c.Load(); err != nil {
		libol.Error("NewConfig.load %s", err)
	}

	libol.Init(c.LogFile, c.Verbose)
	c.Save(fmt.Sprintf("%s.cur", c.SaveFile))

	str, err := libol.Marshal(c, false)
	if err != nil {
		libol.Error("NewConfig.json error: %s", err)
	}
	libol.Debug("NewConfig.json: %s", str)

	return
}

func (c *Config) Default() {
	RightAddr(&c.TcpListen, 10002)
	RightAddr(&c.HttpListen, 10000)

	// TODO reset zero value to default
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

	if c.Links != nil {
		for _, link := range c.Links {
			link.Default()
		}
	}
	return nil
}
