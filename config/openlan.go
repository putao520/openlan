package config

import (
	"flag"
	"fmt"
	"github.com/lightstar-dev/openlan-go/libol"
)

type OpenLan struct {
	Alias      string      `json:"alias"`
	HttpDir    string      `json:"http.dir"`
	HttpListen string      `json:"http.addr"`
	IfMtu      int         `json:"if.mtu"`
	IfAddr     string      `json:"if.addr"`
	Token      string      `json:"admin.token"`
	TokenFile  string      `json:"admin.file"`
	Password   string      `json:"auth.file"`
	Redis      RedisConfig `json:"redis"`
	LogFile    string      `json:"log.file"`
	Verbose    int         `json:"log.level"`
	CrtFile    string      `json:"tls.crt"`
	KeyFile    string      `json:"tls.key"`
	SaveFile   string      `json:"-"`
}

var OpenLanDefault = OpenLan{
	Alias:      "",
	Verbose:    libol.INFO,
	HttpListen: "0.0.0.0:10001",
	Token:      "",
	TokenFile:  ".openlan.token",
	Password:   ".password",
	IfMtu:      1518,
	IfAddr:     "",
	Redis: RedisConfig{
		Addr:   "127.0.0.1",
		Auth:   "",
		Db:     0,
		Enable: false,
	},
	LogFile:  ".openlan.error",
	SaveFile: ".openlan.json",
	CrtFile:  "",
	KeyFile:  "",
	HttpDir:  "public",
}

func NewOpenLan() (c *OpenLan) {
	c = &OpenLan{
		Redis:   OpenLanDefault.Redis,
		LogFile: OpenLanDefault.LogFile,
	}

	flag.StringVar(&c.Alias, "alias", OpenLanDefault.Alias, "the alias for this switch")
	flag.IntVar(&c.Verbose, "log:level", OpenLanDefault.Verbose, "logger level")
	flag.StringVar(&c.LogFile, "log:file", OpenLanDefault.LogFile, "logger file")
	flag.StringVar(&c.HttpListen, "http:addr", OpenLanDefault.HttpListen, "the http listen on")
	flag.StringVar(&c.HttpDir, "http:dir", OpenLanDefault.HttpDir, "the http working directory")
	flag.StringVar(&c.Token, "admin:token", OpenLanDefault.Token, "Administrator token")
	flag.StringVar(&c.TokenFile, "admin:file", OpenLanDefault.TokenFile, "The file administrator token saved to")
	flag.StringVar(&c.Password, "auth:file", OpenLanDefault.Password, "The file password loading from.")
	flag.IntVar(&c.IfMtu, "if:mtu", OpenLanDefault.IfMtu, "the interface MTU include ethernet")
	flag.StringVar(&c.IfAddr, "if:addr", OpenLanDefault.IfAddr, "the interface address")
	flag.StringVar(&c.SaveFile, "conf", OpenLanDefault.SaveFile, "The configuration file")
	flag.StringVar(&c.CrtFile, "tls:crt", OpenLanDefault.CrtFile, "The X509 certificate file for TLS")
	flag.StringVar(&c.KeyFile, "tls:key", OpenLanDefault.KeyFile, "The X509 certificate key for TLS")

	flag.Parse()
	c.Default()
	if err := c.Load(); err != nil {
		libol.Error("NewOpenLan.load %s", err)
	}

	libol.Init(c.LogFile, c.Verbose)
	c.Save(fmt.Sprintf("%s.cur", c.SaveFile))

	str, err := libol.Marshal(c, false)
	if err != nil {
		libol.Error("NewOpenLan.json error: %s", err)
	}
	libol.Debug("NewOpenLan.json: %s", str)

	return
}

func (c *OpenLan) Right() {
	if c.Alias == "" {
		c.Alias = GetAlias()
	}
	RightAddr(&c.HttpListen, 10001)
}

func (c *OpenLan) Default() {
	c.Right()
	// TODO reset zero value to default
}

func (c *OpenLan) Save(file string) error {
	if file == "" {
		file = c.SaveFile
	}

	return libol.MarshalSave(c, file, true)
}

func (c *OpenLan) Load() error {
	if err := libol.UnmarshalLoad(c, c.SaveFile); err != nil {
		return err
	}

	return nil
}

func init() {
	OpenLanDefault.Right()
}
