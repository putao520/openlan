package config

import (
	"flag"
	"github.com/danieldin95/openlan-go/src/libol"
	"runtime"
)

type SocksProxy struct {
	Listen string   `json:"listen,omitempty"`
	Auth   Password `json:"auth,omitempty"`
}

type HttpProxy struct {
	Listen string   `json:"listen,omitempty"`
	Auth   Password `json:"auth,omitempty"`
	Cert   *Cert    `json:"cert,omitempty"`
}

type TcpProxy struct {
	Listen string   `json:"listen,omitempty"`
	Target []string `json:"target,omitempty"`
}

type Proxy struct {
	Conf  string        `json:"-"`
	Log   Log           `json:"log"`
	Socks []*SocksProxy `json:"socks,omitempty"`
	Http  []*HttpProxy  `json:"http,omitempty"`
	Tcp   []*TcpProxy   `json:"tcp,omitempty"`
	PProf string        `json:"pprof"`
}

var xd = &Proxy{
	Log: Log{
		File:    "./openlan-proxy.log",
		Verbose: libol.INFO,
	},
}

func NewProxy() (p Proxy) {
	flag.StringVar(&p.Log.File, "log:file", xd.Log.File, "Configure log file")
	flag.StringVar(&p.Conf, "conf", xd.Conf, "The configure file")
	flag.StringVar(&p.PProf, "prof", xd.PProf, "Http listen for CPU prof")
	flag.IntVar(&p.Log.Verbose, "log:level", xd.Log.Verbose, "Configure log level")
	flag.Parse()
	p.Initialize()
	return p
}

func (p *Proxy) Initialize() {
	if err := p.Load(); err != nil {
		libol.Error("Switch.Initialize %s", err)
	}
	p.Default()
	libol.Debug("Proxy.Initialize %v", p)
}

func (p *Proxy) Right() {
	for _, h := range p.Http {
		if h.Cert != nil {
			h.Cert.Right()
		}
	}
}

func (p *Proxy) Default() {
	p.Right()
}

func (p *Proxy) Load() error {
	return libol.UnmarshalLoad(p, p.Conf)
}

func init() {
	xd.Right()
	if runtime.GOOS == "linux" {
		xd.Log.File = "/var/log/openlan-proxy.log"
	}
}
