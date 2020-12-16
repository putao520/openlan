package config

import (
	"flag"
	"github.com/danieldin95/openlan-go/src/libol"
	"runtime"
	"strings"
)

type Point struct {
	Alias       string    `json:"alias,omitempty"`
	Connection  string    `json:"connection"`
	Timeout     int       `json:"timeout"`
	Username    string    `json:"username,omitempty"`
	Network     string    `json:"network"`
	Password    string    `json:"password,omitempty"`
	Protocol    string    `json:"protocol,omitempty"`
	Interface   Interface `json:"interface"`
	Log         Log       `json:"log"`
	Http        *Http     `json:"http,omitempty"`
	Crypt       *Crypt    `json:"crypt,omitempty"`
	PProf       string    `json:"pprof,omitempty"`
	RequestAddr bool      `json:"-"`
	SaveFile    string    `json:"-"`
	Queue       *Queue    `json:"queue"`
	Terminal    string    `json:"-"`
	Cert        *Cert     `json:"cert"`
}

var defaultPoint = &Point{
	Alias:      "",
	Connection: "openlan.net",
	Network:    "default",
	Protocol:   "tls", // udp, kcp, tcp, tls, ws and wss etc.
	Timeout:    60,
	Log: Log{
		File:    "./openlan-point.log",
		Verbose: libol.INFO,
	},
	Interface: Interface{
		IfMtu:    1518,
		Provider: "kernel",
		Name:     "",
	},
	SaveFile:    "./point.json",
	RequestAddr: true,
	Crypt:       &Crypt{},
	Cert:        &Cert{},
	Terminal:    "on",
}

func NewPoint() *Point {
	pin := &Point{
		RequestAddr: true,
		Crypt:       defaultPoint.Crypt,
		Cert:        defaultPoint.Cert,
	}
	pin.Flags()
	pin.Initialize()
	if Manager.Point == nil {
		Manager.Point = pin
	}
	return pin
}

func (pin *Point) Flags() {
	flag.StringVar(&pin.Alias, "alias", defaultPoint.Alias, "Alias for this point")
	flag.StringVar(&pin.Terminal, "terminal", defaultPoint.Terminal, "Run interactive terminal")
	flag.StringVar(&pin.Connection, "conn", defaultPoint.Connection, "Connection access to")
	flag.StringVar(&pin.Username, "user", defaultPoint.Username, "User access to by <username>@<network>")
	flag.StringVar(&pin.Password, "pass", defaultPoint.Password, "Password for authentication")
	flag.StringVar(&pin.Protocol, "proto", defaultPoint.Protocol, "IP Protocol for connection")
	flag.StringVar(&pin.Log.File, "log:file", defaultPoint.Log.File, "Log saved to file")
	flag.StringVar(&pin.Interface.Name, "if:name", defaultPoint.Interface.Name, "Configure interface name")
	flag.StringVar(&pin.Interface.Address, "if:addr", defaultPoint.Interface.Address, "Configure interface address")
	flag.StringVar(&pin.Interface.Bridge, "if:br", defaultPoint.Interface.Bridge, "Configure bridge name")
	flag.StringVar(&pin.Interface.Provider, "if:provider", defaultPoint.Interface.Provider, "Interface provider")
	flag.StringVar(&pin.SaveFile, "conf", defaultPoint.SaveFile, "The configuration file")
	flag.StringVar(&pin.Crypt.Secret, "crypt:secret", defaultPoint.Crypt.Secret, "Crypt secret")
	flag.StringVar(&pin.Crypt.Algo, "crypt:algo", defaultPoint.Crypt.Algo, "Crypt algorithm")
	flag.StringVar(&pin.PProf, "pprof", defaultPoint.PProf, "Http listen for CPU prof")
	flag.StringVar(&pin.Cert.CaFile, "cacert", defaultPoint.Cert.CaFile, "CA certificate file")
	flag.IntVar(&pin.Timeout, "timeout", defaultPoint.Timeout, "Timeout(s) for socket write/read")
	flag.IntVar(&pin.Log.Verbose, "log:level", defaultPoint.Log.Verbose, "Log level")
	flag.Parse()
}

func (pin *Point) Id() string {
	return pin.Connection + ":" + pin.Network
}

func (pin *Point) Initialize() {
	if err := pin.Load(); err != nil {
		libol.Warn("NewPoint.load %s", err)
	}
	pin.Default()
	libol.SetLogger(pin.Log.File, pin.Log.Verbose)
}

func (pin *Point) Right() {
	if pin.Alias == "" {
		pin.Alias = GetAlias()
	}
	if pin.Network == "" {
		if strings.Contains(pin.Username, "@") {
			pin.Network = strings.SplitN(pin.Username, "@", 2)[1]
		} else {
			pin.Network = defaultPoint.Network
		}
	}
	RightAddr(&pin.Connection, 10002)
	if runtime.GOOS == "darwin" {
		pin.Interface.Provider = "tun"
	}
	if pin.Protocol == "tls" || pin.Protocol == "wss" {
		if pin.Cert == nil {
			pin.Cert = defaultPoint.Cert
		}
	}
	if pin.Cert != nil {
		if pin.Cert.Dir == "" {
			pin.Cert.Dir = "."
		}
		pin.Cert.Right()
	}
}

func (pin *Point) Default() {
	pin.Right()
	if pin.Queue == nil {
		pin.Queue = &Queue{}
	}
	pin.Queue.Default()
	//reset zero value to default
	if pin.Connection == "" {
		pin.Connection = defaultPoint.Connection
	}
	if pin.Interface.IfMtu == 0 {
		pin.Interface.IfMtu = defaultPoint.Interface.IfMtu
	}
	if pin.Timeout == 0 {
		pin.Timeout = defaultPoint.Timeout
	}
	if pin.Crypt != nil {
		pin.Crypt.Default()
	}
}

func (pin *Point) Load() error {
	if err := libol.FileExist(pin.SaveFile); err == nil {
		return libol.UnmarshalLoad(pin, pin.SaveFile)
	}
	return nil
}

func init() {
	defaultPoint.Right()
	if runtime.GOOS == "linux" {
		defaultPoint.Log.File = "/var/log/openlan-point.log"
	} else {
		defaultPoint.Log.File = "./openlan-point.log"
	}
}
