package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/xtaci/kcp-go/v5"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

func VarDir(name ...string) string {
	return "/var/openlan/" + strings.Join(name, "/")
}

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
	QdSrd = 1024 * 4
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

func LogFile(file string) string {
	if runtime.GOOS == "linux" {
		return "/var/log/" + file
	}
	return file
}

type Http struct {
	Listen string `json:"listen,omitempty"`
	Public string `json:"public,omitempty" yaml:"publicDir"`
}

type Crypt struct {
	Algo   string `json:"algo,omitempty" yaml:"algorithm"`
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

type Cert struct {
	Dir      string `json:"dir" yaml:"directory"`
	CrtFile  string `json:"crt" yaml:"cert"`
	KeyFile  string `json:"key" yaml:"key"`
	CaFile   string `json:"ca" yaml:"rootCa"`
	Insecure bool   `json:"insecure"`
}

func (c *Cert) Correct() {
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

type Bridge struct {
	Network  string `json:"network"`
	Peer     string `json:"peer,omitempty" yaml:"peer,omitempty"`
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	IPMtu    int    `json:"mtu,omitempty" yaml:"mtu,omitempty"`
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`
	Stp      string `json:"stp,omitempty" yaml:"stpState,omitempty"`
	Delay    int    `json:"delay,omitempty" yaml:"forwardDelay,omitempty"`
	Mss      int    `json:"mss,omitempty" yaml:"tcpMss,omitempty"`
}

func (br *Bridge) Correct() {
	if br.Name == "" {
		br.Name = "br-" + br.Network
	}
	if br.Provider == "" {
		br.Provider = "linux"
	}
	if br.IPMtu == 0 {
		br.IPMtu = 1500
	}
	if br.Delay == 0 {
		br.Delay = 2
	}
	if br.Stp == "" {
		br.Stp = "on"
	}
}

type IpSubnet struct {
	Network string `json:"network"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Netmask string `json:"netmask"`
}

type MultiPath struct {
	NextHop string `json:"nexthop"`
	Weight  int    `json:"weight"`
}

type PrefixRoute struct {
	Network   string      `json:"network"`
	Prefix    string      `json:"prefix"`
	NextHop   string      `json:"nexthop"`
	MultiPath []MultiPath `json:"multipath"`
	Metric    int         `json:"metric"`
	Mode      string      `json:"mode" yaml:"forwardMode"` // route or snat
}

type HostLease struct {
	Network  string `json:"network"`
	Hostname string `json:"hostname"`
	Address  string `json:"address"`
}

type Password struct {
	Network  string `json:"network"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type FlowRule struct {
	Table    string `json:"table,omitempty" yaml:"table,omitempty"`
	Chain    string `json:"chain,omitempty" yaml:"chain,omitempty"`
	Input    string `json:"input,omitempty" yaml:"input,omitempty"`
	Source   string `json:"source,omitempty" yaml:"source,omitempty"`
	ToSource string `json:"to-source,omitempty" yaml:"toSource,omitempty"`
	Dest     string `json:"destination,omitempty" yaml:"destination,omitempty"`
	ToDest   string `json:"to-destination" yaml:"toDestination,omitempty"`
	Output   string `json:"output,omitempty" yaml:"output,omitempty"`
	Comment  string `json:"comment,omitempty" yaml:"comment,omitempty"`
	Jump     string `json:"jump,omitempty" yaml:"jump,omitempty"` // SNAT/RETURN/MASQUERADE
}

type Interface struct {
	Name     string `json:"name,omitempty"`
	IPMtu    int    `json:"mtu"`
	Address  string `json:"address,omitempty"`
	Bridge   string `json:"bridge,omitempty"`
	Provider string `json:"provider,omitempty"`
	Cost     int    `json:"cost,omitempty"`
}

func CorrectAddr(listen *string, port int) {
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
	case "aes-256":
		block, _ = kcp.NewAESBlockCrypt(pass[:32])
	case "tea":
		block, _ = kcp.NewTEABlockCrypt(pass[:16])
	case "xtea":
		block, _ = kcp.NewXTEABlockCrypt(pass[:16])
	default:
		block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	}
	return block
}
