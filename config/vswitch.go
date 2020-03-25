package config

import (
	"flag"
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"runtime"
)

type Bridge struct {
	Alias    string   `json:"-"`
	Tenant   string   `json:"tenant"`
	IfMtu    int      `json:"if.mtu"`
	IfAddr   string   `json:"if.addr"`
	BrName   string   `json:"if.br"`
	Links    []*Point `json:"links"`
	Bridger  string   `json:"bridger"`
	Password string   `json:"-"`
	Network  string   `json:"-"`
}

type VSwitch struct {
	Alias      string   `json:"alias"`
	TcpListen  string   `json:"vs.addr"`
	HttpDir    string   `json:"http.dir"`
	HttpListen string   `json:"http.addr"`
	Token      string   `json:"admin.token"`
	ConfDir    string   `json:"-"`
	TokenFile  string   `json:"-"`
	SaveFile   string   `json:"-"`
	LogFile    string   `json:"log.file"`
	Verbose    int      `json:"log.level"`
	CrtDir     string   `json:"crt.dir"`
	CrtFile    string   `json:"-"`
	KeyFile    string   `json:"-"`
	Script     string   `json:"script"`
	Bridge     []Bridge `json:"bridge"`
}

var VSwitchDefault = VSwitch{
	Alias:      "",
	Verbose:    libol.INFO,
	HttpListen: "",
	TcpListen:  "0.0.0.0:10002",
	Token:      "",
	ConfDir:    ".",
	LogFile:    "vswitch.log",
	CrtDir:     "",
	HttpDir:    "public",
	Script:     fmt.Sprintf("vswitch.%s.cmd", runtime.GOOS),
}

var BridgeDefault = Bridge{
	Tenant:  "default",
	BrName:  "",
	Bridger: "linux",
	IfMtu:   1518,
}

func NewVSwitch() (c VSwitch) {
	c = VSwitch{
		LogFile: VSwitchDefault.LogFile,
	}
	flag.IntVar(&c.Verbose, "log:level", VSwitchDefault.Verbose, "logger level")
	flag.StringVar(&c.ConfDir, "conf:dir", VSwitchDefault.ConfDir, "The directory configuration on.")
	flag.Parse()
	c.SaveFile = fmt.Sprintf("%s/vswitch.json", c.ConfDir)
	if err := c.Load(); err != nil {
		libol.Error("NewVSwitch.load %s", err)
	}

	c.Default()
	libol.Debug(" %s", c)
	libol.Init(c.LogFile, c.Verbose)
	_ = c.Save(fmt.Sprintf("%s.cur", c.SaveFile))
	libol.Debug("NewVSwitch.json: %v", c)
	return c
}

func (c *VSwitch) Right() {
	if c.Alias == "" {
		c.Alias = GetAlias()
	}
	RightAddr(&c.TcpListen, 10002)
	RightAddr(&c.HttpListen, 10000)

	c.TokenFile = fmt.Sprintf("%s/token", c.ConfDir)
	c.SaveFile = fmt.Sprintf("%s/vswitch.json", c.ConfDir)
	if c.CrtDir != "" {
		c.CrtFile = fmt.Sprintf("%s/crt.pem", c.CrtDir)
		c.KeyFile = fmt.Sprintf("%s/private.key", c.CrtDir)
	}
}

func (c *VSwitch) Default() {
	c.Right()
	if c.Bridge == nil {
		c.Bridge = make([]Bridge, 1)
		c.Bridge[0] = BridgeDefault
	}
	for k := range c.Bridge {
		tenant := c.Bridge[k].Tenant
		if c.Bridge[k].BrName == "" {
			c.Bridge[k].BrName = "br-" + tenant
		}
		c.Bridge[k].Alias = c.Alias
		c.Bridge[k].Network = fmt.Sprintf("%s/network/%s.json", c.ConfDir, tenant)
		c.Bridge[k].Password = fmt.Sprintf("%s/password/%s.json", c.ConfDir, tenant)
		if c.Bridge[k].Bridger == "" {
			c.Bridge[k].Bridger = BridgeDefault.Bridger
		}
		if c.Bridge[k].IfMtu == 0 {
			c.Bridge[k].IfMtu = BridgeDefault.IfMtu
		}
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
	for _, br := range c.Bridge {
		br.Alias = c.Alias
		if br.Links != nil {
			for _, link := range br.Links {
				link.Default()
			}
		}
	}
	return nil
}

func init() {
	VSwitchDefault.Right()
}
