package olsw

import (
	"bytes"
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	OpenVPNBin    = "/usr/sbin/openvpn"
	DefaultCurDir = "/var/openlan/openvpn/default"
)

type OpenVPNData struct {
	Local    string
	Port     string
	Ca       string
	Cert     string
	Key      string
	DhPem    string
	TlsAuth  string
	Cipher   string
	Server   string
	Device   string
	Protocol string
	Script   string
	Routes   []string
	Renego   int
	Stats    string
	IpIp     string
	Push     []string
}

const (
	xAuthConfTmpl = `# Generate by OpenLAN
local {{ .Local }}
port {{ .Port }}
proto {{ .Protocol }}
dev {{ .Device }}
reneg-sec {{ .Renego }}
keepalive 10 120
persist-key
persist-tun
ca {{ .Ca }}
cert {{ .Cert }}
key {{ .Key }}
dh {{ .DhPem }}
server {{ .Server }}
{{- range .Routes }}
push "route {{ . }}"
{{- end }}
{{- range .Push }}
push "{{ . }}"
{{- end }}
ifconfig-pool-persist {{ .Protocol }}{{ .Port }}ipp
tls-auth {{ .TlsAuth }} 0
cipher {{ .Cipher }}
status {{ .Protocol }}{{ .Port }}server.status 5
verify-client-cert none
script-security 3
auth-user-pass-verify "{{ .Script }}" via-env
username-as-common-name
verb 3
`
	certConfTmpl = `# Generate by OpenLAN
local {{ .Local }}
port {{ .Port }}
proto {{ .Protocol }}
dev {{ .Device }}
reneg-sec {{ .Renego }}
keepalive 10 120
persist-key
persist-tun
ca {{ .Ca }}
cert {{ .Cert }}
key {{ .Key }}
dh {{ .DhPem }}
server {{ .Server }}
{{- range .Routes }}
push "route {{ . }}"
{{- end }}
ifconfig-pool-persist {{ .Protocol }}{{ .Port }}ipp
tls-auth {{ .TlsAuth }} 0
cipher {{ .Cipher }}
status {{ .Protocol }}{{ .Port }}server.status 5
verb 3
`
)

func NewOpenVpnDataFromConf(obj *OpenVPN) *OpenVPNData {
	cfg := obj.Cfg
	data := &OpenVPNData{
		Local:    obj.Local,
		Port:     obj.Port,
		Ca:       cfg.RootCa,
		Cert:     cfg.ServerCrt,
		Key:      cfg.ServerKey,
		DhPem:    cfg.DhPem,
		TlsAuth:  cfg.TlsAuth,
		Cipher:   cfg.Cipher,
		Device:   cfg.Device,
		Protocol: cfg.Protocol,
		Script:   cfg.Script,
		Renego:   cfg.Renego,
		Push:     cfg.Push,
	}
	addr, _ := libol.IPNetwork(cfg.Subnet)
	data.Server = strings.ReplaceAll(addr, "/", " ")
	for _, rt := range cfg.Routes {
		if addr, err := libol.IPNetwork(rt); err == nil {
			r := strings.ReplaceAll(addr, "/", " ")
			data.Routes = append(data.Routes, r)
		}
	}
	return data
}

type OpenVPN struct {
	Cfg      *config.OpenVPN
	out      *libol.SubLogger
	Protocol string
	Local    string
	Port     string
}

func NewOpenVPN(cfg *config.OpenVPN) *OpenVPN {
	obj := &OpenVPN{
		Cfg:      cfg,
		out:      libol.NewSubLogger(cfg.Network),
		Protocol: cfg.Protocol,
		Local:    "0.0.0.0",
		Port:     "4494",
	}
	obj.Local = strings.SplitN(cfg.Listen, ":", 2)[0]
	if strings.Contains(cfg.Listen, ":") {
		obj.Port = strings.SplitN(cfg.Listen, ":", 2)[1]
	}
	return obj
}

func (o *OpenVPN) ID() string {
	return o.Protocol + o.Port
}

func (o *OpenVPN) Path() string {
	return OpenVPNBin
}

func (o *OpenVPN) Directory() string {
	if o.Cfg == nil {
		return DefaultCurDir
	}
	return o.Cfg.Directory
}

func (o *OpenVPN) ServerConf() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, o.ID()+"server.conf")
}

func (o *OpenVPN) ClientConf() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, o.ID()+"client.ovpn")
}

func (o *OpenVPN) ServerLog() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, o.ID()+"server.log")
}

func (o *OpenVPN) ServerPid() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, o.ID()+"server.pid")
}

func (o *OpenVPN) ServerStats() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, o.ID()+"server.stats")
}

func (o *OpenVPN) ServerTmpl() string {
	tmplStr := xAuthConfTmpl
	if o.Cfg.Auth == "cert" {
		tmplStr = certConfTmpl
	}
	cfgTmpl := filepath.Join(o.Cfg.Directory, o.ID()+"server.tmpl")
	_ = ioutil.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
	return tmplStr
}

func (o *OpenVPN) IppTxt() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, o.ID()+"ipp")
}

func (o *OpenVPN) WriteConf(path string) error {
	fp, err := libol.CreateFile(path)
	if err != nil || fp == nil {
		return err
	}
	defer fp.Close()
	data := NewOpenVpnDataFromConf(o)
	o.out.Debug("OpenVPN.WriteConf %v", data)
	tmplStr := o.ServerTmpl()
	if tmpl, err := template.New("main").Parse(tmplStr); err != nil {
		return err
	} else {
		if err := tmpl.Execute(fp, data); err != nil {
			return err
		}
	}
	return nil
}

func (o *OpenVPN) Clean() {
	files := []string{o.ServerStats(), o.IppTxt()}
	for _, file := range files {
		if err := libol.FileExist(file); err == nil {
			if err := os.Remove(file); err != nil {
				o.out.Warn("OpenVPN.Clean %s", err)
			}
		}
	}
}

func (o *OpenVPN) Initialize() {
	if !o.ValidConf() {
		return
	}
	o.Clean()
	if err := os.Mkdir(o.Directory(), 0600); err != nil {
		o.out.Warn("OpenVPN.Initialize %s", err)
	}
	if err := o.WriteConf(o.ServerConf()); err != nil {
		o.out.Warn("OpenVPN.Initialize %s", err)
		return
	}
	if ctx, err := o.Profile(); err == nil {
		file := o.ClientConf()
		if err := ioutil.WriteFile(file, ctx, 0600); err != nil {
			o.out.Warn("OpenVPN.Initialize %s", err)
		}
	} else {
		o.out.Warn("OpenVPN.Initialize %s", err)
	}
}

func (o *OpenVPN) ValidConf() bool {
	if o.Cfg == nil {
		return false
	}
	if o.Cfg.Listen == "" || o.Cfg.Subnet == "" {
		return false
	}
	return true
}

func (o *OpenVPN) Start() {
	if !o.ValidConf() {
		return
	}
	log, err := libol.CreateFile(o.ServerLog())
	if err != nil {
		o.out.Warn("OpenVPN.Start %s", err)
		return
	}
	libol.Go(func() {
		defer log.Close()
		args := []string{
			"--cd", o.Directory(),
			"--config", o.ServerConf(),
			"--writepid", o.ServerPid(),
		}
		cmd := exec.Command(o.Path(), args...)
		cmd.Stdout = log
		cmd.Stderr = log
		if err := cmd.Run(); err != nil {
			o.out.Error("OpenVPN.Start %s: %s", o.ID(), err)
		}
	})
}

func (o *OpenVPN) Stop() {
	if !o.ValidConf() {
		return
	}
	if data, err := ioutil.ReadFile(o.ServerPid()); err != nil {
		o.out.Debug("OpenVPN.Stop %s", err)
	} else {
		pid := strings.TrimSpace(string(data))
		cmd := exec.Command("/usr/bin/kill", pid)
		if err := cmd.Run(); err != nil {
			o.out.Warn("OpenVPN.Stop %s: %s", pid, err)
		}
	}
	o.Clean()
}

func (o *OpenVPN) ProfileTmpl() string {
	tmplStr := xAuthClientProfile
	if o.Cfg.Auth == "cert" {
		tmplStr = certClientProfile
	}
	cfgTmpl := filepath.Join(o.Cfg.Directory, o.ID()+"client.tmpl")
	_ = ioutil.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
	return tmplStr
}

func (o *OpenVPN) Profile() ([]byte, error) {
	data := NewOpenVpnProfileFromConf(o)
	tmplStr := o.ProfileTmpl()
	tmpl, err := template.New("main").Parse(tmplStr)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err == nil {
		return out.Bytes(), nil
	} else {
		return nil, err
	}
}

type OpenVPNProfile struct {
	Server   string
	Port     string
	Ca       string
	Cert     string
	Key      string
	TlsAuth  string
	Cipher   string
	Device   string
	Protocol string
	Renego   int
}

const (
	xAuthClientProfile = `# Generate by OpenLAN
client
dev {{ .Device }}
route-metric 300
proto {{ .Protocol }}
remote {{ .Server }} {{ .Port }}
reneg-sec {{ .Renego }}
resolv-retry infinite
nobind
persist-key
persist-tun
<ca>
{{ .Ca -}}
</ca>
remote-cert-tls server
<tls-auth>
{{ .TlsAuth -}}
</tls-auth>
key-direction 1
cipher {{ .Cipher }}
auth-nocache
verb 4
auth-user-pass
`
	certClientProfile = `# Generate by OpenLAN
client
dev {{ .Device }}
route-metric 300
proto {{ .Protocol }}
remote {{ .Server }} {{ .Port }}
reneg-sec {{ .Renego }}
resolv-retry infinite
nobind
persist-key
persist-tun
<ca>
{{ .Ca -}}
</ca>
remote-cert-tls server
<tls-auth>
{{ .TlsAuth -}}
</tls-auth>
key-direction 1
cipher {{ .Cipher }}
auth-nocache
verb 4
`
)

func NewOpenVpnProfileFromConf(obj *OpenVPN) *OpenVPNProfile {
	cfg := obj.Cfg
	data := &OpenVPNProfile{
		Server:   obj.Local,
		Port:     obj.Port,
		Cipher:   cfg.Cipher,
		Device:   cfg.Device[:3],
		Protocol: cfg.Protocol,
		Renego:   cfg.Renego,
	}
	if ctx, err := ioutil.ReadFile(cfg.RootCa); err == nil {
		data.Ca = string(ctx)
	}
	if ctx, err := ioutil.ReadFile(cfg.TlsAuth); err == nil {
		data.TlsAuth = string(ctx)
	}
	return data
}
