package config

import "strings"

type OpenVPN struct {
	Network   string   `json:"network"`
	Directory string   `json:"directory"`
	Listen    string   `json:"listen"`
	Protocol  string   `json:"protocol"`
	Subnet    string   `json:"subnet"`
	Device    string   `json:"device"`
	Auth      string   `json:"auth"` // xauth or cert.
	DhPem     string   `json:"dh"`
	RootCa    string   `json:"ca"`
	ServerCrt string   `json:"cert"`
	ServerKey string   `json:"key"`
	TlsAuth   string   `json:"tlsAuth"`
	Cipher    string   `json:"cipher"`
	Routes    []string `json:"routes"`
	Script    string   `json:"-"`
}

func DefaultOpenVPN() *OpenVPN {
	return &OpenVPN{
		Protocol:  "tcp",
		Auth:      "xauth",
		Device:    "tun",
		RootCa:    "/var/openlan/cert/ca.crt",
		ServerCrt: "/var/openlan/cert/crt",
		ServerKey: "/var/openlan/cert/key",
		DhPem:     "/var/openlan/openvpn/dh.pem",
		TlsAuth:   "/var/openlan/openvpn/ta.key",
		Cipher:    "AES-256-CBC",
		Script:    "/usr/bin/openlan",
	}
}

func (o *OpenVPN) Right(obj *OpenVPN) {
	if o.Directory == "" {
		o.Directory = "/var/openlan/openvpn/" + o.Network
	}
	if o.Auth == "" && obj != nil {
		o.Auth = obj.Auth
	}
	if o.Device == "" {
		if strings.Contains(o.Listen, ":") {
			o.Device = "tun-" + strings.SplitN(o.Listen, ":", 2)[1]
		} else {
			o.Device = "tun"
		}
	}
	if o.Protocol == "" && obj != nil {
		o.Protocol = obj.Protocol
	}
	if o.DhPem == "" && obj != nil {
		o.DhPem = obj.DhPem
	}
	if o.RootCa == "" && obj != nil {
		o.RootCa = obj.RootCa
	}
	if o.ServerCrt == "" && obj != nil {
		o.ServerCrt = obj.ServerCrt
	}
	if o.ServerKey == "" && obj != nil {
		o.ServerKey = obj.ServerKey
	}
	if o.TlsAuth == "" && obj != nil {
		o.TlsAuth = obj.TlsAuth
	}
	if o.Cipher == "" && obj != nil {
		o.Cipher = obj.Cipher
	}
	if obj != nil {
		bin := obj.Script + " user check --network " + o.Network
		o.Script = bin
	}
}
