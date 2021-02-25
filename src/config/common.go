package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/xtaci/kcp-go/v5"
	"io/ioutil"
	"os"
	"strings"
)

type Queue struct {
	SockWr int `json:"swr"` // per frames about 1572(1514+4+20+20+14)bytes
	SockRd int `json:"srd"` // per frames
	TapWr  int `json:"twr"` // per frames about 1572((1514+4+20+20+14))bytes
	TapRd  int `json:"trd"` // per frames
	VirSnd int `json:"vsd"`
	VirWrt int `json:"vwr"`
}

var (
	QdSwr = 1024 * 4
	QdSrd = 8
	QdTwr = 1024 * 2
	QdTrd = 2
	QdVsd = 1024 * 8
	QdVWr = 1024 * 4
)

func (q *Queue) Default() {
	if q.SockWr == 0 {
		q.SockWr = QdSwr
	}
	if q.SockRd == 0 {
		q.SockRd = QdSrd
	}
	if q.TapWr == 0 {
		q.TapWr = QdTwr
	}
	if q.TapRd == 0 {
		q.TapRd = QdTrd
	}
	if q.VirSnd == 0 {
		q.VirSnd = QdVsd
	}
	if q.VirWrt == 0 {
		q.VirWrt = QdVWr
	}
	libol.Debug("Queue.Default %v", q)
}

type Log struct {
	File    string `json:"file,omitempty"`
	Verbose int    `json:"level,omitempty"`
}

type Http struct {
	Listen string `json:"listen,omitempty"`
	Public string `json:"public,omitempty"`
}

type Crypt struct {
	Algo   string `json:"algo,omitempty"`
	Secret string `json:"secret,omitempty"`
}

func (c *Crypt) IsZero() bool {
	return c.Algo == "" && c.Secret == ""
}

func (c *Crypt) Default() {
	if c.Secret != "" && c.Algo == "" {
		c.Algo = "xor"
	}
}

func RightAddr(listen *string, port int) {
	values := strings.Split(*listen, ":")
	if len(values) == 1 {
		*listen = fmt.Sprintf("%s:%d", values[0], port)
	}
}

func GetAlias() string {
	if hostname, err := os.Hostname(); err == nil {
		return strings.ToLower(hostname)
	}
	return libol.GenRandom(13)
}

type Cert struct {
	Dir      string `json:"dir"`
	CrtFile  string `json:"crt"`
	KeyFile  string `json:"key"`
	CaFile   string `json:"ca"`
	Insecure bool   `json:"insecure"`
}

func (c *Cert) Right() {
	if c.Dir == "" {
		return
	}
	if c.CrtFile == "" {
		c.CrtFile = fmt.Sprintf("%s/crt", c.Dir)
	}
	if c.KeyFile == "" {
		c.KeyFile = fmt.Sprintf("%s/key", c.Dir)
	}
	if c.CaFile == "" {
		c.CaFile = fmt.Sprintf("%s/ca-trusted.crt", c.Dir)
	}
}

func (c *Cert) GetTlsCfg() *tls.Config {
	if c.KeyFile == "" || c.CrtFile == "" {
		return nil
	}
	libol.Debug("Cert.GetTlsCfg: %v", c)
	cer, err := tls.LoadX509KeyPair(c.CrtFile, c.KeyFile)
	if err != nil {
		libol.Error("Cert.GetTlsCfg: %s", err)
		return nil
	}
	return &tls.Config{Certificates: []tls.Certificate{cer}}
}

func (c *Cert) GetCertPool() *x509.CertPool {
	if c.CaFile == "" {
		return nil
	}
	if err := libol.FileExist(c.CaFile); err != nil {
		libol.Debug("Cert.GetTlsCertPool: %s not such file", c.CaFile)
		return nil
	}
	caCert, err := ioutil.ReadFile(c.CaFile)
	if err != nil {
		libol.Warn("Cert.GetTlsCertPool: %s", err)
		return nil
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		libol.Warn("Cert.GetTlsCertPool: invalid cert")
	}
	return pool
}

func GetBlock(cfg *Crypt) kcp.BlockCrypt {
	if cfg == nil || cfg.IsZero() {
		return nil
	}
	var block kcp.BlockCrypt
	pass := make([]byte, 64)
	if len(cfg.Secret) <= 64 {
		copy(pass, cfg.Secret)
	} else {
		copy(pass, []byte(cfg.Secret)[:64])
	}
	switch cfg.Algo {
	case "aes-128":
		block, _ = kcp.NewAESBlockCrypt(pass[:16])
	case "aes-192":
		block, _ = kcp.NewAESBlockCrypt(pass[:24])
	default:
		block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	}
	return block
}

type Bridge struct {
	Peer     string `json:"peer"`
	Name     string `json:"name"`
	IfMtu    int    `json:"mtu"`
	Address  string `json:"address,omitempty"`
	Provider string `json:"provider"`
	Stp      string `json:"stp"`
	Delay    int    `json:"delay"`
}

type IpSubnet struct {
	Start   string `json:"start"`
	End     string `json:"end"`
	Netmask string `json:"netmask"`
}

type PrefixRoute struct {
	Prefix  string `json:"prefix"`
	NextHop string `json:"nexthop"`
	Metric  int    `json:"metric"`
	Mode    string `json:"mode"` // route or snat
}

type HostLease struct {
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
}

type Password struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type OpenVPN struct {
	Name      string   `json:"-"`
	WorkDir   string   `json:"-"`
	Listen    string   `json:"listen"`
	Protocol  string   `json:"protocol"`
	Subnet    string   `json:"subnet"`
	Device    string   `json:"device"`
	Auth      string   `json:"auth"` // xauth or cert.
	DhPem     string   `json:"dh"`
	RootCa    string   `json:"ca"`
	ServerCrt string   `json:"cert"`
	ServerKey string   `json:"key"`
	TlsAuth   string   `json:"tlsauth"`
	Cipher    string   `json:"cipher"`
	Routes    []string `json:"routes"`
	Script    string   `json:"-"`
}

var defaultOvpn = OpenVPN{
	Protocol:  "tcp",
	Auth:      "xauth",
	Device:    "tun",
	RootCa:    "/var/openlan/cert/ca.crt",
	ServerCrt: "/var/openlan/cert/crt",
	ServerKey: "/var/openlan/cert/key",
	DhPem:     "/var/openlan/openvpn/dh.pem",
	TlsAuth:   "/var/openlan/openvpn/ta.key",
	Cipher:    "AES-256-CBC",
	Script:    "/usr/bin/openlan-pass",
}

func (o *OpenVPN) Right() {
	if o.WorkDir == "" {
		o.WorkDir = "/var/openlan/openvpn/" + o.Name
	}
	if o.Auth == "" {
		o.Auth = defaultOvpn.Auth
	}
	if o.Device == "" {
		o.Device = defaultOvpn.Device
	}
	if o.Protocol == "" {
		o.Protocol = defaultOvpn.Protocol
	}
	if o.DhPem == "" {
		o.DhPem = defaultOvpn.DhPem
	}
	if o.RootCa == "" {
		o.RootCa = defaultOvpn.RootCa
	}
	if o.ServerCrt == "" {
		o.ServerCrt = defaultOvpn.ServerCrt
	}
	if o.ServerKey == "" {
		o.ServerKey = defaultOvpn.ServerKey
	}
	if o.TlsAuth == "" {
		o.TlsAuth = defaultOvpn.TlsAuth
	}
	if o.Cipher == "" {
		o.Cipher = defaultOvpn.Cipher
	}
	bin := defaultOvpn.Script
	bin += " " + strings.Join(os.Args[1:], " ")
	bin += " -zone " + o.Name
	o.Script = bin
}

type Network struct {
	Alias    string        `json:"-"`
	Name     string        `json:"name,omitempty"`
	Bridge   Bridge        `json:"bridge,omitempty"`
	Subnet   IpSubnet      `json:"subnet,omitempty"`
	OpenVPN  *OpenVPN      `json:"openvpn,omitempty"`
	Links    []*Point      `json:"links,omitempty"`
	Hosts    []HostLease   `json:"hosts,omitempty"`
	Routes   []PrefixRoute `json:"routes,omitempty"`
	Password []Password    `json:"password,omitempty"`
}

func (n *Network) Right() {
	if n.Bridge.Name == "" {
		n.Bridge.Name = "br-" + n.Name
	}
	if n.Bridge.Provider == "" {
		n.Bridge.Provider = "linux"
	}
	if n.Bridge.IfMtu == 0 {
		n.Bridge.IfMtu = 1518
	}
	if n.Bridge.Delay == 0 {
		n.Bridge.Delay = 2
	}
	if n.Bridge.Stp == "" {
		n.Bridge.Stp = "on"
	}
	ifAddr := strings.SplitN(n.Bridge.Address, "/", 2)[0]
	for i := range n.Routes {
		if n.Routes[i].Metric == 0 {
			n.Routes[i].Metric = 592
		}
		if n.Routes[i].NextHop == "" {
			n.Routes[i].NextHop = ifAddr
		}
		if n.Routes[i].Mode == "" {
			n.Routes[i].Mode = "snat"
		}
	}
	if n.OpenVPN != nil {
		n.OpenVPN.Name = n.Name
		n.OpenVPN.Right()
	}
}

type FlowRule struct {
	Table    string `json:"table"`
	Chain    string `json:"chain"`
	Input    string `json:"input"`
	Source   string `json:"source"`
	ToSource string `json:"to-source"`
	Dest     string `json:"destination"`
	ToDest   string `json:"to-destination"`
	Output   string `json:"output"`
	Comment  string `json:"comment"`
	Jump     string `json:"jump"` // SNAT/RETURN/MASQUERADE
}

var defaultPerf = Perf{
	Point:    1024,
	Neighbor: 1024,
	OnLine:   64,
	Link:     1024,
	User:     1024,
}

type Perf struct {
	Point    int `json:"point"`
	Neighbor int `json:"neighbor"`
	OnLine   int `json:"online"`
	Link     int `json:"link"`
	User     int `json:"user"`
}

func (p *Perf) Right() {
	if p.Point == 0 {
		p.Point = defaultPerf.Point
	}
	if p.Neighbor == 0 {
		p.Neighbor = defaultPerf.Neighbor
	}
	if p.OnLine == 0 {
		p.OnLine = defaultPerf.OnLine
	}
	if p.Link == 0 {
		p.Link = defaultPerf.Link
	}
	if p.User == 0 {
		p.User = defaultPerf.User
	}
}

type Interface struct {
	Name     string `json:"name,omitempty"`
	IfMtu    int    `json:"mtu"`
	Address  string `json:"address,omitempty"`
	Bridge   string `json:"bridge,omitempty"`
	Provider string `json:"provider,omitempty"`
	Cost     int    `json:"cost,omitempty"`
}

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
