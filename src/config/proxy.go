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

func DefaultProxy() *Proxy {
	obj := &Proxy{
		Log: Log{
			File:    "./openlan-proxy.log",
			Verbose: libol.INFO,
		},
	}
	obj.Right(nil)
	if runtime.GOOS == "linux" {
		obj.Log.File = "/var/log/openlan-proxy.log"
	} else {
		obj.Log.File = "./openlan-proxy.log"
	}
	return obj
}

func NewProxy() *Proxy {
	px := &Proxy{}
	px.Flags()
	px.Parse()
	px.Initialize()
	if Manager.Proxy == nil {
		Manager.Proxy = px
	}
	return px
}

func (p *Proxy) Flags() {
	obj := DefaultProxy()
	flag.StringVar(&p.Log.File, "log:file", obj.Log.File, "Configure log file")
	flag.StringVar(&p.Conf, "conf", obj.Conf, "The configure file")
	flag.StringVar(&p.PProf, "prof", obj.PProf, "Http listen for CPU prof")
	flag.IntVar(&p.Log.Verbose, "log:level", obj.Log.Verbose, "Configure log level")
}

func (p *Proxy) Parse() {
	flag.Parse()
}

func (p *Proxy) Initialize() {
	if err := p.Load(); err != nil {
		libol.Error("Switch.Initialize %s", err)
	}
	p.Default()
	libol.Debug("Proxy.Initialize %v", p)
}

func (p *Proxy) Right(obj *Proxy) {
	for _, h := range p.Http {
		if h.Cert != nil {
			h.Cert.Right()
		}
	}
}

func (p *Proxy) Default() {
	p.Right(nil)
}

func (p *Proxy) Load() error {
	return libol.UnmarshalLoad(p, p.Conf)
}
