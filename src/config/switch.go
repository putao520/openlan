package config

import (
	"flag"
	"github.com/danieldin95/openlan-go/src/libol"
	"path/filepath"
	"runtime"
)

func DefaultPerf() *Perf {
	return &Perf{
		Point:    1024,
		Neighbor: 1024,
		OnLine:   64,
		Link:     1024,
		User:     1024,
	}
}

type Perf struct {
	Point    int `json:"point"`
	Neighbor int `json:"neighbor"`
	OnLine   int `json:"online"`
	Link     int `json:"link"`
	User     int `json:"user"`
}

func (p *Perf) Right(obj *Perf) {
	if p.Point == 0 && obj != nil {
		p.Point = obj.Point
	}
	if p.Neighbor == 0 && obj != nil {
		p.Neighbor = obj.Neighbor
	}
	if p.OnLine == 0 && obj != nil {
		p.OnLine = obj.OnLine
	}
	if p.Link == 0 && obj != nil {
		p.Link = obj.Link
	}
	if p.User == 0 && obj != nil {
		p.User = obj.User
	}
}

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

func DefaultSwitch() *Switch {
	obj := &Switch{
		Timeout: 120,
		Log: Log{
			File:    "./openlan-switch.log",
			Verbose: libol.INFO,
		},
		Http: &Http{
			Listen: "0.0.0.0:10000",
		},
		Listen: "0.0.0.0:10002",
		Perf:   DefaultPerf(),
	}
	obj.Right(nil)
	if runtime.GOOS == "linux" {
		obj.Log.File = "/var/log/openlan-switch.log"
	} else {
		obj.Log.File = "./openlan-switch.log"
	}
	return obj
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
	obj := DefaultSwitch()
	flag.StringVar(&sw.Log.File, "log:file", obj.Log.File, "Configure log file")
	flag.StringVar(&sw.ConfDir, "conf:dir", obj.ConfDir, "Configure switch's directory")
	flag.StringVar(&sw.PProf, "prof", obj.PProf, "Http listen for CPU prof")
	flag.IntVar(&sw.Log.Verbose, "log:level", obj.Log.Verbose, "Configure log level")
}

func (sw *Switch) Parse() {
	flag.Parse()
}

func (sw *Switch) Initialize() {
	sw.SaveFile = sw.ConfDir + "/switch.json"
	if err := sw.Load(); err != nil {
		libol.Error("Switch.Initialize %s", err)
	}
	sw.Default()
	libol.Debug("Switch.Initialize %v", sw)
}

func (sw *Switch) Right(obj *Switch) {
	if sw.Alias == "" {
		sw.Alias = GetAlias()
	}
	RightAddr(&sw.Listen, 10002)
	if sw.Http != nil {
		RightAddr(&sw.Http.Listen, 10000)
	}
	libol.Debug("Proxy.Right Http %v", sw.Http)
	sw.TokenFile = sw.ConfDir + "/token"
	sw.SaveFile = sw.ConfDir + "/switch.json"
	if sw.Cert != nil {
		sw.Cert.Right()
		// default is tls if cert configured
		if sw.Protocol == "" {
			sw.Protocol = "tls"
		}
	}
	if sw.Perf != nil {
		obj := DefaultPerf()
		sw.Perf.Right(obj)
	} else {
		sw.Perf = DefaultPerf()
	}
	if sw.Password == "" {
		sw.Password = sw.ConfDir + "/password"
	}
}

func (sw *Switch) Default() {
	obj := DefaultSwitch()
	sw.Right(obj)
	if sw.Network == nil {
		sw.Network = make([]*Network, 0, 32)
	}
	if sw.Timeout == 0 {
		sw.Timeout = obj.Timeout
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
