package olsw

import (
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
	Local   string
	Port    string
	Ca      string
	Cert    string
	Key     string
	DhPem   string
	TlsAuth string
	Cipher  string
	Server  string
	Dev     string
	Proto   string
	Script  string
	Routes  []string
}

func NewOpenVpnDataFromConf(cfg *config.OpenVPN) *OpenVPNData {
	data := &OpenVPNData{
		Local:   strings.SplitN(cfg.Listen, ":", 2)[0],
		Ca:      cfg.RootCa,
		Cert:    cfg.ServerCrt,
		Key:     cfg.ServerKey,
		DhPem:   cfg.DhPem,
		TlsAuth: cfg.TlsAuth,
		Cipher:  cfg.Cipher,
		Dev:     cfg.Device,
		Proto:   cfg.Protocol,
		Script:  cfg.Script,
	}
	if addr, err := libol.IPNetwork(cfg.Subnet); err == nil {
		data.Server = strings.ReplaceAll(addr, "/", " ")
	}
	if strings.Contains(cfg.Listen, ":") {
		data.Port = strings.SplitN(cfg.Listen, ":", 2)[1]
	}
	for _, rt := range cfg.Routes {
		if addr, err := libol.IPNetwork(rt); err == nil {
			data.Routes = append(data.Routes, strings.ReplaceAll(addr, "/", " "))
		}
	}
	return data
}

const (
	xAuthConfTmpl = `# OpenVPN configuration
local {{ .Local }}
port {{ .Port }}
proto {{ .Proto }}
dev {{ .Dev }}
ca {{ .Ca }}
cert {{ .Cert }}
key {{ .Key }}
dh {{ .DhPem }}
server {{ .Server }}
{{ range .Routes }}
push "route {{ . }}"
{{ end }}
ifconfig-pool-persist ipp.txt
keepalive 10 120
tls-auth {{ .TlsAuth }} 0
cipher {{ .Cipher }}
persist-key
persist-tun
status status.log
verify-client-cert none
script-security 3
auth-user-pass-verify "{{ .Script }}" via-env
username-as-common-name
verb 3
`
	certConfTmpl = `# OpenVPN configuration
local {{ .Local }}
port {{ .Port }}
proto {{ .Proto }}
dev {{ .Dev }}
ca {{ .Ca }}
cert {{ .Cert }}
key {{ .Key }}
dh {{ .DhPem }}
server {{ .Server }}
# Push routes to the client
{{ range .Routes }}
push "route {{ . }}"
{{ end }}
ifconfig-pool-persist ipp.txt
keepalive 10 120
tls-auth {{ .TlsAuth }} 0
cipher {{ .Cipher }}
persist-key
persist-tun
status status.log
verb 3
`
)

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
	if o.Cfg.Auth != "xauth" {
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
