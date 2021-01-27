package config

import (
	"flag"
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"path/filepath"
	"runtime"
)

type Switch struct {
	Alias     string     `json:"alias"`
	Perf      *Perf      `json:"perf,omitempty"`
	Protocol  string     `json:"protocol"` // tcp, tls, udp, kcp, ws and wss.
	Listen    string     `json:"listen"`
	Timeout   int        `json:"timeout"`
	Http      *Http      `json:"http,omitempty"`
	Log       Log        `json:"log"`
	Cert      *Cert      `json:"cert,omitempty"`
	Crypt     *Crypt     `json:"crypt,omitempty"`
	PProf     string     `json:"pprof"`
	Network   []*Network `json:"network,omitempty"`
	FireWall  []FlowRule `json:"firewall,omitempty"`
	Inspect   []string   `json:"inspect"`
	Queue     *Queue     `json:"queue"`
	Password  string     `json:"password"`
	ConfDir   string     `json:"-"`
	TokenFile string     `json:"-"`
	SaveFile  string     `json:"-"`
}

var defaultSwitch = &Switch{
	Timeout: 120,
	Log: Log{
		File:    "./openlan-switch.log",
		Verbose: libol.INFO,
	},
	Http: &Http{
		Listen: "0.0.0.0:10000",
	},
	Listen: "0.0.0.0:10002",
	Perf:   &defaultPerf,
}

func NewSwitch() *Switch {
	sw := &Switch{}
	sw.Flags()
	sw.Parse()
	sw.Initialize()
	if Manager.Switch == nil {
		Manager.Switch = sw
	}
	return sw
}

func (sw *Switch) Flags() {
	flag.StringVar(&sw.Log.File, "log:file", defaultSwitch.Log.File, "Configure log file")
	flag.StringVar(&sw.ConfDir, "conf:dir", defaultSwitch.ConfDir, "Configure switch's directory")
	flag.StringVar(&sw.PProf, "prof", defaultSwitch.PProf, "Http listen for CPU prof")
	flag.IntVar(&sw.Log.Verbose, "log:level", defaultSwitch.Log.Verbose, "Configure log level")
}

func (sw *Switch) Parse() {
	flag.Parse()
}

func (sw *Switch) Initialize() {
	sw.SaveFile = fmt.Sprintf("%s/switch.json", sw.ConfDir)
	if err := sw.Load(); err != nil {
		libol.Error("Switch.Initialize %s", err)
	}
	sw.Default()
	libol.Debug("Switch.Initialize %v", sw)
}

func (sw *Switch) Right() {
	if sw.Alias == "" {
		sw.Alias = GetAlias()
	}
	RightAddr(&sw.Listen, 10002)
	if sw.Http != nil {
		RightAddr(&sw.Http.Listen, 10000)
	}
	libol.Debug("Proxy.Right Http %v", sw.Http)
	sw.TokenFile = fmt.Sprintf("%s/token", sw.ConfDir)
	sw.SaveFile = fmt.Sprintf("%s/switch.json", sw.ConfDir)
	if sw.Cert != nil {
		sw.Cert.Right()
		// default is tls if cert configured
		if sw.Protocol == "" {
			sw.Protocol = "tls"
		}
	}
	if sw.Perf != nil {
		sw.Perf.Right()
	} else {
		sw.Perf = &defaultPerf
	}
	if sw.Password == "" {
		sw.Password = fmt.Sprintf("%s/password", sw.ConfDir)
	}
}

func (sw *Switch) Default() {
	sw.Right()
	if sw.Network == nil {
		sw.Network = make([]*Network, 0, 32)
	}
	if sw.Timeout == 0 {
		sw.Timeout = defaultSwitch.Timeout
	}
	if sw.Crypt != nil {
		sw.Crypt.Default()
	}
	if sw.Queue == nil {
		sw.Queue = &Queue{}
	}
	sw.Queue.Default()
	files, err := filepath.Glob(sw.ConfDir + "/network/*.json")
	if err != nil {
		libol.Error("Switch.Default %s", err)
	}
	for _, k := range files {
		n := &Network{
			Alias: sw.Alias,
		}
		if err := libol.UnmarshalLoad(n, k); err != nil {
			libol.Error("Switch.Default %s", err)
			continue
		}
		sw.Network = append(sw.Network, n)
	}
	for _, n := range sw.Network {
		for _, link := range n.Links {
			link.Default()
		}
		n.Right()
		n.Alias = sw.Alias
	}
}

func (sw *Switch) Load() error {
	return libol.UnmarshalLoad(sw, sw.SaveFile)
}

func init() {
	defaultSwitch.Right()
	if runtime.GOOS == "linux" {
		defaultSwitch.Log.File = "/var/log/openlan-switch.log"
	} else {
		defaultSwitch.Log.File = "./openlan-switch.log"
	}
}
