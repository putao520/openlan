package config

import (
	"flag"
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"path/filepath"
)

type Bridge struct {
	Name     string `json:"name"`
	Mtu      int    `json:"mtu"`
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Provider string `json:"provider"`
}

type IpRange struct {
	Start   string `json:"start"`
	End     string `json:"end"`
	Netmask string `json:"netmask"`
}

type IpRoute struct {
	Prefix  string `json:"prefix"`
	Nexthop string `json:"nexthop"`
}

type IpSet struct {
	Route []IpRoute `json:"routes"`
	Range IpRange   `json:"range"`
}

type Password struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Network struct {
	Alias    string     `json:"-"`
	Name     string     `json:"name" yaml:"name"`
	Bridge   Bridge     `json:"bridge" yaml:"bridge"`
	Links    []*Point   `json:"links" yaml:"links"`
	IpSet    *IpSet     `json:"ip,omitempty" yaml:"ip,omitempty"`
	Password []Password `json:"password"`
}

func (n *Network) Right() {
	if n.Bridge.Name == "" {
		n.Bridge.Name = "br-" + n.Name
	}
	if n.Bridge.Provider == "" {
		n.Bridge.Provider = "linux"
	}
	if n.Bridge.Mtu == 0 {
		n.Bridge.Mtu = 1518
	}
}

type Cert struct {
	Dir     string `json:"dir"`
	CrtFile string `json:"crt" yaml:"crt"`
	KeyFile string `json:"key" yaml:"key"`
}

type FlowRules struct {
	Table    string `json:"table"`
	Chain    string `json:"chain"`
	Input    string `json:"input"`
	Source   string `json:"source"`
	ToSource string `json:"to-source"`
	Dest     string `json:"destination"`
	ToDest   string `json:"to-destination"`
	Output   string `json:"output"`
	Comment  string `json:"comment"`
	Jump     string `json:"jump"` // SNAT/RETURN/MASQUERADE
}

type VSwitch struct {
	Alias     string      `json:"alias"`
	Listen    string      `json:"listen"`
	Http      *Http       `json:"http,omitempty" yaml:"http,omitempty"`
	Log       Log         `json:"log" yaml:"log"`
	Cert      Cert        `json:"cert"`
	Network   []*Network  `json:"network"`
	FireWall  []FlowRules `json:"firewall"`
	ConfDir   string      `json:"-" yaml:"-"`
	TokenFile string      `json:"-" yaml:"-"`
	SaveFile  string      `json:"-" yaml:"-"`
}

var vSwitchDef = VSwitch{
	Alias: "",
	Log: Log{
		File:    "./openlan-vswitch.log",
		Verbose: libol.INFO,
	},
	Http: &Http{
		Listen: "0.0.0.0:10000",
	},
	Listen: "0.0.0.0:10002",
}

func NewVSwitch() (c VSwitch) {
	flag.IntVar(&c.Log.Verbose, "log:level", vSwitchDef.Log.Verbose, "Configure log level")
	flag.StringVar(&c.ConfDir, "conf:dir", vSwitchDef.ConfDir, "Configure virtual switch directory")
	flag.Parse()

	c.SaveFile = fmt.Sprintf("%s/vswitch.json", c.ConfDir)
	if err := c.Load(); err != nil {
		libol.Error("NewVSwitch.load %s", err)
	}
	c.Default()
	libol.Init(c.Log.File, c.Log.Verbose)
	libol.Debug("NewVSwitch %v", c)
	return c
}

func (c *VSwitch) Right() {
	if c.Alias == "" {
		c.Alias = GetAlias()
	}
	RightAddr(&c.Listen, 10002)
	if c.Http != nil {
		RightAddr(&c.Http.Listen, 10000)
	}
	c.TokenFile = fmt.Sprintf("%s/token", c.ConfDir)
	c.SaveFile = fmt.Sprintf("%s/vswitch.json", c.ConfDir)
	if c.Cert.Dir != "" {
		c.Cert.CrtFile = fmt.Sprintf("%s/crt.pem", c.Cert.Dir)
		c.Cert.KeyFile = fmt.Sprintf("%s/private.key", c.Cert.Dir)
	}
}

func (c *VSwitch) Default() {
	c.Right()
	if c.Network == nil {
		c.Network = make([]*Network, 0, 32)
	}

	files, err := filepath.Glob(c.ConfDir + "/network/*.json")
	if err != nil {
		libol.Error("VSwitch.Default %s", err)
	}
	for _, k := range files {
		n := &Network{
			Alias: c.Alias,
		}
		if err := libol.UnmarshalLoad(n, k); err != nil {
			libol.Error("VSwitch.Default %s", err)
			continue
		}
		c.Network = append(c.Network, n)
	}
	for _, n := range c.Network {
		for _, link := range n.Links {
			link.Default()
		}
		n.Right()
		n.Alias = c.Alias
	}
}

func (c *VSwitch) Load() error {
	return libol.UnmarshalLoad(c, c.SaveFile)
}

func init() {
	vSwitchDef.Right()
}
