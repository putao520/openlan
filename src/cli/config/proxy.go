package config

import (
	"flag"
	"github.com/danieldin95/openlan-go/src/libol"
	"runtime"
)

type Proxy struct {
	Conf  string        `json:"-"`
	Log   Log           `json:"log"`
	Socks []*SocksProxy `json:"socks,omitempty"`
	Http  []*HttpProxy  `json:"http,omitempty"`
	Tcp   []*TcpProxy   `json:"tcp,omitempty"`
	PProf string        `json:"pprof"`
}

var defaultProxy = &Proxy{
	Log: Log{
		File:    "./openlan-proxy.log",
		Verbose: libol.INFO,
	},
}

func NewProxy() *Proxy {
	px := &Proxy{}
	px.Flags()
	px.Initialize()
	if Manager.Proxy == nil {
		Manager.Proxy = px
	}
	return px
}

func (px *Proxy) Flags() {
	flag.StringVar(&px.Log.File, "log:file", defaultProxy.Log.File, "Configure log file")
	flag.StringVar(&px.Conf, "conf", defaultProxy.Conf, "The configure file")
	flag.StringVar(&px.PProf, "prof", defaultProxy.PProf, "Http listen for CPU prof")
	flag.IntVar(&px.Log.Verbose, "log:level", defaultProxy.Log.Verbose, "Configure log level")
	flag.Parse()
}
func (px *Proxy) Initialize() {
	if err := px.Load(); err != nil {
		libol.Error("Switch.Initialize %s", err)
	}
	px.Default()
	libol.Debug("Proxy.Initialize %v", px)
}

func (px *Proxy) Right() {
	for _, h := range px.Http {
		if h.Cert != nil {
			h.Cert.Right()
		}
	}
}

func (px *Proxy) Default() {
	px.Right()
}

func (px *Proxy) Load() error {
	return libol.UnmarshalLoad(px, px.Conf)
}

func init() {
	defaultProxy.Right()
	if runtime.GOOS == "linux" {
		defaultProxy.Log.File = "/var/log/openlan-proxy.log"
	} else {
		defaultProxy.Log.File = "./openlan-proxy.log"
	}
}
