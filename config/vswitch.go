package config

import (
	"flag"
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"runtime"
)

type VSwitch struct {
	Alias      string   `json:"alias"`
	TcpListen  string   `json:"vs.addr"`
	HttpDir    string   `json:"http.dir"`
	HttpListen string   `json:"http.addr"`
	IfMtu      int      `json:"if.mtu"`
	IfAddr     string   `json:"if.addr"`
	BrName     string   `json:"if.br"`
	Token      string   `json:"admin.token"`
	ConfDir    string   `json:"conf.dir"`
	TokenFile  string   `json:"-"`
	Password   string   `json:"-"`
	Network    string   `json:"-"`
	SaveFile   string   `json:"-"`
	LogFile    string   `json:"log.file"`
	Verbose    int      `json:"log.level"`
	CrtDir     string   `json:"crt.dir"`
	CrtFile    string   `json:"-"`
	KeyFile    string   `json:"-"`
	Links      []*Point `json:"links"`
	Script     string   `json:"script"`
}

var VSwitchDefault = VSwitch{
	Alias:      "",
	BrName:     "",
	Verbose:    libol.INFO,
	HttpListen: "",
	TcpListen:  "0.0.0.0:10002",
	Token:      "",
	ConfDir:    ".",
	IfMtu:      1518,
	IfAddr:     "",
	LogFile:    "vswitch.error",
	CrtDir:     "",
	HttpDir:    "public",
	Links:      nil,
	Script:     fmt.Sprintf("vswitch.%s.cmd", runtime.GOOS),
}

func NewVSwitch() (c *VSwitch) {
	c = &VSwitch{
		LogFile: VSwitchDefault.LogFile,
	}

	flag.StringVar(&c.Alias, "alias", VSwitchDefault.Alias, "the alias for this switch")
	flag.IntVar(&c.Verbose, "log:level", VSwitchDefault.Verbose, "logger level")
	flag.StringVar(&c.LogFile, "log:file", VSwitchDefault.LogFile, "logger file")
	flag.StringVar(&c.HttpListen, "http:addr", VSwitchDefault.HttpListen, "the http listen on")
	flag.StringVar(&c.HttpDir, "http:dir", VSwitchDefault.HttpDir, "the http working directory")
	flag.StringVar(&c.TcpListen, "vs:addr", VSwitchDefault.TcpListen, "the server listen on")
	flag.StringVar(&c.Token, "admin:token", VSwitchDefault.Token, "Administrator token")
	flag.StringVar(&c.ConfDir, "conf:dir", VSwitchDefault.ConfDir, "The directory configuration on.")
	flag.IntVar(&c.IfMtu, "if:mtu", VSwitchDefault.IfMtu, "the interface MTU include ethernet")
	flag.StringVar(&c.IfAddr, "if:addr", VSwitchDefault.IfAddr, "the interface address")
	flag.StringVar(&c.BrName, "if:br", VSwitchDefault.BrName, "the bridge name")
	flag.StringVar(&c.CrtDir, "crt:dir", VSwitchDefault.CrtFile, "The directory X509 certificate file on")
	flag.StringVar(&c.Script, "script", VSwitchDefault.Script, "call script you assigned")

	flag.Parse()
	c.SaveFile = fmt.Sprintf("%s/vswitch.json", c.ConfDir)
	if err := c.Load(); err != nil {
		libol.Error("NewVSwitch.load %s", err)
	}
	c.Default()

	libol.Debug(" %s", c)

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

	c.TokenFile = fmt.Sprintf("%s/token", c.ConfDir)
	c.Password = fmt.Sprintf("%s/password", c.ConfDir)
	c.SaveFile = fmt.Sprintf("%s/vswitch.json", c.ConfDir)
	c.Network = fmt.Sprintf("%s/network.json", c.ConfDir)

	if c.CrtDir != "" {
		c.CrtFile = fmt.Sprintf("%s/crt.pem", c.CrtDir)
		c.KeyFile = fmt.Sprintf("%s/private.key", c.CrtDir)
	}
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
