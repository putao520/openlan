package config

import (
	"fmt"
	"strings"
)

type OpenVPN struct {
	Network   string     `json:"network"`
	Directory string     `json:"directory"`
	Listen    string     `json:"listen"`
	Protocol  string     `json:"protocol"`
	Subnet    string     `json:"subnet"`
	Device    string     `json:"device"`
	Auth      string     `json:"auth"` // xauth or cert.
	DhPem     string     `json:"dh"`
	RootCa    string     `json:"ca"`
	ServerCrt string     `json:"cert"`
	ServerKey string     `json:"key"`
	TlsAuth   string     `json:"tlsAuth"`
	Cipher    string     `json:"cipher"`
	Routes    []string   `json:"routes"`
	Renego    int        `json:"renego"`
	Script    string     `json:"-"`
	Breed     []*OpenVPN `json:"breed"`
	Push      []string   `json:"push"`
}

var index = 1194

func genTunName() string {
	index += 1
	return fmt.Sprintf("tun-%d", index)
}

func DefaultOpenVPN() *OpenVPN {
	return &OpenVPN{
		Protocol:  "tcp",
		Auth:      "xauth",
		Device:    "tun0",
		RootCa:    VarDir("cert/ca.crt"),
		ServerCrt: VarDir("cert/crt"),
		ServerKey: VarDir("cert/key"),
		DhPem:     VarDir("openvpn/dh.pem"),
		TlsAuth:   VarDir("openvpn/ta.key"),
		Cipher:    "AES-256-CBC",
		Script:    "/usr/bin/openlan",
	}
}

func (o *OpenVPN) Correct(obj *OpenVPN) {
	if obj != nil {
		if o.Network == "" {
			o.Network = obj.Network
		}
		if o.Auth == "" {
			o.Auth = obj.Auth
		}
		if o.Protocol == "" {
			o.Protocol = obj.Protocol
		}
		if o.DhPem == "" {
			o.DhPem = obj.DhPem
		}
		if o.RootCa == "" {
			o.RootCa = obj.RootCa
		}
		if o.ServerCrt == "" {
			o.ServerCrt = obj.ServerCrt
		}
		if o.ServerKey == "" {
			o.ServerKey = obj.ServerKey
		}
		if o.TlsAuth == "" {
			o.TlsAuth = obj.TlsAuth
		}
		if o.Cipher == "" {
			o.Cipher = obj.Cipher
		}
		if o.Routes == nil || len(o.Routes) == 0 {
			o.Routes = append(o.Routes, obj.Routes...)
		}
		if o.Push == nil || len(o.Push) == 0 {
			o.Push = append(o.Push, obj.Push...)
		}
		if o.Script == "" {
			bin := obj.Script + " user check --network " + o.Network
			o.Script = bin
		}
	}
	if o.Directory == "" {
		o.Directory = VarDir("openvpn", o.Network)
	}
	if o.Device == "" {
		if !strings.Contains(o.Listen, ":") {
			o.Listen += ":1194"
		}
		o.Device = genTunName()
	}
	for _, ch := range o.Breed {
		ch.Correct(o)
	}
}
