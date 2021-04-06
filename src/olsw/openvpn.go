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
}

const (
	xAuthConfTmpl = `# Generate by OpenLAN
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
	certConfTmpl = `# Generate by OpenLAN
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
		Port:     "1194",
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

type OpenVpn struct {
	Cfg *config.OpenVPN
	out *libol.SubLogger
}

func NewOpenVPN(cfg *config.OpenVPN) *OpenVpn {
	return &OpenVpn{
		Cfg: cfg,
		out: libol.NewSubLogger(cfg.Network),
	}
}

func (o *OpenVpn) Path() string {
	return OpenVPNBin
}

func (o *OpenVpn) Directory() string {
	if o.Cfg == nil {
		return DefaultCurDir
	}
	return o.Cfg.Directory
}

func (o *OpenVpn) ServerConf() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, "server.conf")
}

func (o *OpenVpn) ClientConf() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, "client.ovpn")
}

func (o *OpenVpn) ServerLog() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, "server.log")
}

func (o *OpenVpn) ServerPid() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, "server.pid")
}

func (o *OpenVpn) ServerStats() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, "server.stats")
}

func (o *OpenVpn) ServerTmpl() string {
	tmplStr := xAuthConfTmpl
	if o.Cfg.Auth == "cert" {
		tmplStr = certConfTmpl
	}
	cfgTmpl := filepath.Join(o.Cfg.Directory, "server.tmpl")
	if err := libol.FileExist(cfgTmpl); err == nil {
		if data, err := ioutil.ReadFile(cfgTmpl); err == nil {
			tmplStr = string(data)
		}
	} else {
		_ = ioutil.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
	}
	return tmplStr
}

func (o *OpenVpn) IppTxt() string {
	if o.Cfg == nil {
		return ""
	}
	return filepath.Join(o.Cfg.Directory, "ipp.txt")
}

func (o *OpenVpn) WriteConf(path string) error {
	fp, err := libol.CreateFile(path)
	if err != nil || fp == nil {
		return err
	}
	defer fp.Close()
	data := NewOpenVpnDataFromConf(o.Cfg)
	o.out.Debug("OpenVpn.WriteConf %v", data)
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

func (o *OpenVpn) Clean() {
	status := o.ServerStats()
	if err := libol.FileExist(status); err == nil {
		if err := os.Remove(status); err != nil {
			o.out.Warn("OpenVpn.Clean %s", err)
		}
	}
	ipp := o.IppTxt()
	if err := libol.FileExist(ipp); err == nil {
		if err := os.Remove(ipp); err != nil {
			o.out.Warn("OpenVpn.Clean %s", err)
		}
	}
}

func (o *OpenVpn) Initialize() {
	if !o.ValidConf() {
		return
	}
	o.Clean()
	if err := os.Mkdir(o.Directory(), 0600); err != nil {
		o.out.Warn("OpenVpn.Initialize %s", err)
	}
	if err := o.WriteConf(o.ServerConf()); err != nil {
		o.out.Warn("OpenVpn.Initialize %s", err)
		return
	}
	if ctx, err := o.Profile(); err == nil {
		file := o.ClientConf()
		if err := ioutil.WriteFile(file, ctx, 0600); err != nil {
			o.out.Warn("OpenVpn.Initialize %s", err)
		}
	} else {
		o.out.Warn("OpenVpn.Initialize %s", err)
	}
}

func (o *OpenVpn) ValidConf() bool {
	if o.Cfg == nil {
		return false
	}
	if o.Cfg.Listen == "" || o.Cfg.Subnet == "" {
		return false
	}
	return true
}

func (o *OpenVpn) Start() {
	if !o.ValidConf() {
		return
	}
	log, err := libol.CreateFile(o.ServerLog())
	if err != nil {
		o.out.Warn("OpenVpn.Start %s", err)
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
		if err := cmd.Run(); err != nil {
			o.out.Error("OpenVpn.Start %s", err)
		}
	})
}

func (o *OpenVpn) Stop() {
	if !o.ValidConf() {
		return
	}
	if data, err := ioutil.ReadFile(o.ServerPid()); err != nil {
		o.out.Debug("OpenVpn.Stop %s", err)
	} else {
		pid := strings.TrimSpace(string(data))
		cmd := exec.Command("/usr/bin/kill", pid)
		if err := cmd.Run(); err != nil {
			o.out.Warn("OpenVpn.Stop %s: %s", pid, err)
		}
	}
	o.Clean()
}

func (o *OpenVpn) ProfileTmpl() string {
	tmplStr := xAuthClientProfile
	if o.Cfg.Auth == "cert" {
		tmplStr = certClientProfile
	}
	cfgTmpl := filepath.Join(o.Cfg.Directory, "client.tmpl")
	if err := libol.FileExist(cfgTmpl); err == nil {
		if data, err := ioutil.ReadFile(cfgTmpl); err == nil {
			tmplStr = string(data)
		}
	} else {
		_ = ioutil.WriteFile(cfgTmpl, []byte(tmplStr), 0600)
	}
	return tmplStr
}

func (o *OpenVpn) Profile() ([]byte, error) {
	data := NewOpenVpnProfileFromConf(o.Cfg)
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
	xAuthClientProfile = `# Generate by OpenLAN
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
	certClientProfile = `# Generate by OpenLAN
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
		Device:   cfg.Device[:3],
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
