package config

import (
	"flag"
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"runtime"
)

type VSwitch struct {
	Alias      string      `json:"alias"`
	TcpListen  string      `json:"vs.addr"`
	HttpDir    string      `json:"http.dir"`
	HttpListen string      `json:"http.addr"`
	IfMtu      int         `json:"if.mtu"`
	IfAddr     string      `json:"if.addr"`
	BrName     string      `json:"if.br"`
	Token      string      `json:"admin.token"`
	TokenFile  string      `json:"admin.file"`
	Password   string      `json:"auth.file"`
	Redis      RedisConfig `json:"redis"`
	LogFile    string      `json:"log.file"`
	Verbose    int         `json:"log.level"`
	CrtFile    string      `json:"tls.crt"`
	KeyFile    string      `json:"tls.key"`
	Links      []*Point    `json:"links"`
	Script     string      `json:"script"`

	SaveFile   string      `json:"-"`
}

var VSwitchDefault = VSwitch{
	Alias:      "",
	BrName:     "",
	Verbose:    libol.INFO,
	HttpListen: "",
	TcpListen:  "0.0.0.0:10002",
	Token:      "",
	TokenFile:  "vswitch.token",
	Password:   "password",
	IfMtu:      1518,
	IfAddr:     "",
	Redis: RedisConfig{
		Addr:   "127.0.0.1",
		Auth:   "",
		Db:     0,
		Enable: false,
	},
	LogFile:  "vswitch.error",
	SaveFile: "vswitch.json",
	CrtFile:  "",
	KeyFile:  "",
	HttpDir:  "public",
	Links:    nil,
	Script:   fmt.Sprintf("vswitch.%s.cmd", runtime.GOOS),
}

func NewVSwitch() (c *VSwitch) {
	c = &VSwitch{
		Redis:   VSwitchDefault.Redis,
		LogFile: VSwitchDefault.LogFile,
	}

	flag.StringVar(&c.Alias, "alias", VSwitchDefault.Alias, "the alias for this switch")
	flag.IntVar(&c.Verbose, "log:level", VSwitchDefault.Verbose, "logger level")
	flag.StringVar(&c.LogFile, "log:file", VSwitchDefault.LogFile, "logger file")
	flag.StringVar(&c.HttpListen, "http:addr", VSwitchDefault.HttpListen, "the http listen on")
	flag.StringVar(&c.HttpDir, "http:dir", VSwitchDefault.HttpDir, "the http working directory")
	flag.StringVar(&c.TcpListen, "vs:addr", VSwitchDefault.TcpListen, "the server listen on")
	flag.StringVar(&c.Token, "admin:token", VSwitchDefault.Token, "Administrator token")
	flag.StringVar(&c.TokenFile, "admin:file", VSwitchDefault.TokenFile, "The file administrator token saved to")
	flag.StringVar(&c.Password, "auth:file", VSwitchDefault.Password, "The file password loading from.")
	flag.IntVar(&c.IfMtu, "if:mtu", VSwitchDefault.IfMtu, "the interface MTU include ethernet")
	flag.StringVar(&c.IfAddr, "if:addr", VSwitchDefault.IfAddr, "the interface address")
	flag.StringVar(&c.BrName, "if:br", VSwitchDefault.BrName, "the bridge name")
	flag.StringVar(&c.SaveFile, "conf", VSwitchDefault.SaveFile, "The configuration file")
	flag.StringVar(&c.CrtFile, "tls:crt", VSwitchDefault.CrtFile, "The X509 certificate file for TLS")
	flag.StringVar(&c.KeyFile, "tls:key", VSwitchDefault.KeyFile, "The X509 certificate key for TLS")
	flag.StringVar(&c.Script, "script", VSwitchDefault.Script, "call script you assigned")

	flag.Parse()
	c.Default()
	if err := c.Load(); err != nil {
		libol.Error("NewVSwitch.load %s", err)
	}

	libol.Init(c.LogFile, c.Verbose)
	c.Save(fmt.Sprintf("%s.cur", c.SaveFile))

	str, err := libol.Marshal(c, false)
	if err != nil {
		libol.Error("NewVSwitch.json error: %s", err)
	}
	libol.Debug("NewVSwitch.json: %s", str)

	return
}

func (c *VSwitch) Right() {
	if c.Alias == "" {
		c.Alias = GetAlias()
	}
	RightAddr(&c.TcpListen, 10002)
	RightAddr(&c.HttpListen, 10000)
}

func (c *VSwitch) Default() {
	c.Right()

	if c.IfMtu == 0 {
		c.IfMtu = VSwitchDefault.IfMtu
	}
}

func (c *VSwitch) Save(file string) error {
	if file == "" {
		file = c.SaveFile
	}

	return libol.MarshalSave(c, file, true)
}

func (c *VSwitch) Load() error {
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

func init() {
	VSwitchDefault.Right()
}
