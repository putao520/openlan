package olsw

import (
	"bytes"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"io/ioutil"
	"os"
	"os/exec"
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
}

const (
	xAuthConfTmpl = `# OpenVPN configuration
local {{ .Local }}
port {{ .Port }}
proto {{ .Protocol }}
dev {{ .Device }}
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
ifconfig-pool-persist ipp.txt
tls-auth {{ .TlsAuth }} 0
cipher {{ .Cipher }}
status server.status 5
verify-client-cert none
script-security 3
auth-user-pass-verify "{{ .Script }}" via-env
username-as-common-name
verb 3
`
	certConfTmpl = `# OpenVPN configuration
local {{ .Local }}
port {{ .Port }}
proto {{ .Protocol }}
dev {{ .Device }}
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
ifconfig-pool-persist ipp.txt
tls-auth {{ .TlsAuth }} 0
cipher {{ .Cipher }}
status server.status 5
verb 3
`
)

func NewOpenVpnDataFromConf(cfg *config.OpenVPN) *OpenVPNData {
	data := &OpenVPNData{
		Local:    strings.SplitN(cfg.Listen, ":", 2)[0],
		Ca:       cfg.RootCa,
		Cert:     cfg.ServerCrt,
		Key:      cfg.ServerKey,
		DhPem:    cfg.DhPem,
		TlsAuth:  cfg.TlsAuth,
		Cipher:   cfg.Cipher,
		Device:   cfg.Device,
		Protocol: cfg.Protocol,
		Script:   cfg.Script,
	}
	addr, _ := libol.IPNetwork(cfg.Subnet)
	data.Server = strings.ReplaceAll(addr, "/", " ")
	if strings.Contains(cfg.Listen, ":") {
		data.Port = strings.SplitN(cfg.Listen, ":", 2)[1]
	}
	for _, rt := range cfg.Routes {
		if addr, err := libol.IPNetwork(rt); err == nil {
			r := strings.ReplaceAll(addr, "/", " ")
			data.Routes = append(data.Routes, r)
		}
	}
	return data
}

type OpenVPN struct {
	Cfg *config.OpenVPN
	out *libol.SubLogger
}

func NewOpenVPN(cfg *config.OpenVPN) *OpenVPN {
	return &OpenVPN{
		Cfg: cfg,
		out: libol.NewSubLogger(cfg.Name),
	}
}

func (o *OpenVPN) Path() string {
	return OpenVPNBin
}

func (o *OpenVPN) WorkDir() string {
	if o.Cfg == nil {
		return DefaultCurDir
	}
	return o.Cfg.WorkDir
}

func (o *OpenVPN) ConfFile() string {
	if o.Cfg == nil {
		return ""
	}
	return o.Cfg.WorkDir + "/server.conf"
}

func (o *OpenVPN) ProfileFile() string {
	if o.Cfg == nil {
		return ""
	}
	return o.Cfg.WorkDir + "/client.ovpn"
}

func (o *OpenVPN) LogFile() string {
	if o.Cfg == nil {
		return ""
	}
	return o.Cfg.WorkDir + "/server.log"
}

func (o *OpenVPN) PidFile() string {
	if o.Cfg == nil {
		return ""
	}
	return o.Cfg.WorkDir + "/server.pid"
}

func (o *OpenVPN) WriteConf(path string) error {
	fp, err := libol.CreateFile(path)
	if err != nil || fp == nil {
		return err
	}
	defer fp.Close()
	data := NewOpenVpnDataFromConf(o.Cfg)
	o.out.Debug("OpenVPN.WriteConf %v", data)
	tmplStr := xAuthConfTmpl
	if o.Cfg.Auth == "cert" {
		tmplStr = certConfTmpl
	}
	if tmpl, err := template.New("main").Parse(tmplStr); err != nil {
		return err
	} else {
		if err := tmpl.Execute(fp, data); err != nil {
			return err
		}
	}
	return nil
}

func (o *OpenVPN) Initialize() {
	if !o.ValidCfg() {
		return
	}
	if err := os.Mkdir(o.WorkDir(), 0600); err != nil {
		o.out.Warn("OpenVPN.Initialize %s", err)
	}
	if err := o.WriteConf(o.ConfFile()); err != nil {
		o.out.Warn("OpenVPN.Initialize %s", err)
		return
	}
	if ctx, err := o.Profile(); err == nil {
		file := o.ProfileFile()
		if err := ioutil.WriteFile(file, ctx, 0600); err != nil {
			o.out.Warn("OpenVPN.Initialize %s", err)
		}
	} else {
		o.out.Warn("OpenVPN.Initialize %s", err)
	}
}

func (o *OpenVPN) ValidCfg() bool {
	if o.Cfg == nil {
		return false
	}
	if o.Cfg.Listen == "" || o.Cfg.Subnet == "" {
		return false
	}
	return true
}

func (o *OpenVPN) Start() {
	if !o.ValidCfg() {
		return
	}
	log, err := libol.CreateFile(o.LogFile())
	if err != nil {
		o.out.Warn("OpenVPN.Start %s", err)
		return
	}
	libol.Go(func() {
		defer log.Close()
		args := []string{
			"--cd", o.WorkDir(), "--config", o.ConfFile(), "--writepid", o.PidFile(),
		}
		cmd := exec.Command(o.Path(), args...)
		cmd.Stdout = log
		if err := cmd.Run(); err != nil {
			o.out.Error("OpenVPN.Start %s, and see log %s", err, o.LogFile())
		}
	})
}

func (o *OpenVPN) Stop() {
	if !o.ValidCfg() {
		return
	}
	if data, err := ioutil.ReadFile(o.PidFile()); err != nil {
		o.out.Debug("OpenVPN.Stop %s", err)
		return
	} else {
		pid := strings.TrimSpace(string(data))
		cmd := exec.Command("/usr/bin/kill", pid)
		if err := cmd.Run(); err != nil {
			o.out.Warn("OpenVPN.Stop %s: %s", pid, err)
		}
	}
}

type OpenVPNProfile struct {
	Remote   string
	Ca       string
	Cert     string
	Key      string
	TlsAuth  string
	Cipher   string
	Device   string
	Protocol string
}

const (
	xAuthClientProfile = `
client
dev {{ .Device }}
route-metric 300
proto {{ .Protocol }}
remote {{ .Remote }}
resolv-retry infinite
nobind
persist-key
persist-tun
<ca>
{{ .Ca }}
</ca>
remote-cert-tls server
<tls-auth>
{{ .TlsAuth }}
</tls-auth>
key-direction 1
cipher {{ .Cipher }}
auth-nocache
verb 4
auth-user-pass
`
	certClientProfile = `
client
dev {{ .Device }}
route-metric 300
proto {{ .Protocol }}
remote {{ .Remote }}
resolv-retry infinite
nobind
persist-key
persist-tun
<ca>
{{ .Ca }}
</ca>
<cert>
</cert>
<key>
</key>
remote-cert-tls server
<tls-auth>
{{ .TlsAuth }}
</tls-auth>
key-direction 1
cipher {{ .Cipher }}
auth-nocache
verb 4
`
)

func NewOpenVpnProfileFromConf(cfg *config.OpenVPN) *OpenVPNProfile {
	data := &OpenVPNProfile{
		Remote:   strings.ReplaceAll(cfg.Listen, ":", " "),
		Cipher:   cfg.Cipher,
		Device:   cfg.Device,
		Protocol: cfg.Protocol,
	}
	if ctx, err := ioutil.ReadFile(cfg.RootCa); err == nil {
		data.Ca = string(ctx)
	}
	if ctx, err := ioutil.ReadFile(cfg.TlsAuth); err == nil {
		data.TlsAuth = string(ctx)
	}
	return data
}

func (o *OpenVPN) Profile() ([]byte, error) {
	data := NewOpenVpnProfileFromConf(o.Cfg)
	tmplStr := xAuthClientProfile
	if o.Cfg.Auth == "cert" {
		tmplStr = certClientProfile
	}
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
