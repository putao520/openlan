package config

import (
	"flag"
	"github.com/danieldin95/openlan-go/src/libol"
	"path/filepath"
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

func (p *Perf) Correct(obj *Perf) {
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
	Perf      Perf       `json:"perf,omitempty"`
	Protocol  string     `json:"protocol"` // tcp, tls, udp, kcp, ws and wss.
	Listen    string     `json:"listen"`
	Timeout   int        `json:"timeout"`
	Http      *Http      `json:"http,omitempty"`
	Log       Log        `json:"log"`
	Cert      *Cert      `json:"cert,omitempty"`
	Crypt     *Crypt     `json:"crypt,omitempty"`
	Network   []*Network `json:"network,omitempty"`
	Acl       []*ACL     `json:"acl"`
	FireWall  []FlowRule `json:"firewall,omitempty"`
	Inspect   []string   `json:"inspect"`
	Queue     Queue      `json:"queue"`
	Password  string     `json:"password"`
	Ldap      *LDAP      `json:"ldap"`
	ConfDir   string     `json:"-"`
	TokenFile string     `json:"-"`
	SaveFile  string     `json:"-"`
}

func DefaultSwitch() *Switch {
	obj := &Switch{
		Timeout: 120,
		Log: Log{
			File:    LogFile("openlan-switch.log"),
			Verbose: libol.INFO,
		},
		Http: &Http{
			Listen: "0.0.0.0:10000",
		},
		Listen: "0.0.0.0:10002",
	}
	obj.Correct(nil)
	return obj
}

func NewSwitch() *Switch {
	s := &Switch{}
	s.Flags()
	s.Parse()
	s.Initialize()
	if Manager.Switch == nil {
		Manager.Switch = s
	}
	return s
}

func (s *Switch) Flags() {
	obj := DefaultSwitch()
	flag.StringVar(&s.Log.File, "log:file", obj.Log.File, "Configure log file")
	flag.StringVar(&s.ConfDir, "conf:dir", obj.ConfDir, "Configure switch's directory")
	flag.IntVar(&s.Log.Verbose, "log:level", obj.Log.Verbose, "Configure log level")
}

func (s *Switch) Parse() {
	flag.Parse()
}

func (s *Switch) Initialize() {
	s.SaveFile = filepath.Join(s.ConfDir, "switch.json")
	if err := s.Load(); err != nil {
		libol.Error("Switch.Initialize %s", err)
	}
	s.Default()
	libol.Debug("Switch.Initialize %v", s)
}

func (s *Switch) Correct(obj *Switch) {
	if s.Alias == "" {
		s.Alias = GetAlias()
	}
	CorrectAddr(&s.Listen, 10002)
	if s.Http != nil {
		CorrectAddr(&s.Http.Listen, 10000)
	}
	libol.Debug("Proxy.Correct Http %v", s.Http)
	s.TokenFile = filepath.Join(s.ConfDir, "token")
	s.SaveFile = filepath.Join(s.ConfDir, "switch.json")
	if s.Cert != nil {
		s.Cert.Correct()
	}
	perf := &s.Perf
	perf.Correct(DefaultPerf())
	if s.Password == "" {
		s.Password = filepath.Join(s.ConfDir, "password")
	}
	if s.Protocol == "" {
		s.Protocol = "tcp"
	}
}

func (s *Switch) LoadNetwork() {
	files, err := filepath.Glob(filepath.Join(s.ConfDir, "network", "*.json"))
	if err != nil {
		libol.Error("Switch.LoadNetwork %s", err)
	}
	for _, k := range files {
		obj := &Network{
			Alias: s.Alias,
		}
		if err := libol.UnmarshalLoad(obj, k); err != nil {
			libol.Error("Switch.LoadNetwork %s", err)
			continue
		}
		switch obj.Provider {
		case "esp":
			obj.Interface = &ESPInterface{}
		case "vxlan":
			obj.Interface = &VxLANInterface{}
		}
		if obj.Interface != nil {
			if err := libol.UnmarshalLoad(obj, k); err != nil {
				libol.Error("Switch.LoadNetwork %s", err)
				continue
			}
		}
		s.Network = append(s.Network, obj)
	}
	for _, obj := range s.Network {
		for _, link := range obj.Links {
			link.Default()
		}
		obj.Correct()
		obj.Alias = s.Alias
		obj.Crypt = s.Crypt
	}
}

func (s *Switch) LoadAcl() {
	files, err := filepath.Glob(filepath.Join(s.ConfDir, "acl", "*.json"))
	if err != nil {
		libol.Error("Switch.LoadAcl %s", err)
	}
	for _, k := range files {
		obj := &ACL{}
		if err := libol.UnmarshalLoad(obj, k); err != nil {
			libol.Error("Switch.LoadAcl %s", err)
			continue
		}
		s.Acl = append(s.Acl, obj)
	}
	for _, obj := range s.Acl {
		for _, rule := range obj.Rules {
			rule.Correct()
		}
	}
}

func (s *Switch) Default() {
	obj := DefaultSwitch()
	s.Correct(obj)
	if s.Network == nil {
		s.Network = make([]*Network, 0, 32)
	}
	if s.Timeout == 0 {
		s.Timeout = obj.Timeout
	}
	if s.Crypt != nil {
		s.Crypt.Default()
	}
	queue := &s.Queue
	queue.Default()
	s.LoadAcl()
	s.LoadNetwork()
}

func (s *Switch) Load() error {
	return libol.UnmarshalLoad(s, s.SaveFile)
}
