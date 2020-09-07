package config

import (
	"crypto/tls"
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/xtaci/kcp-go/v5"
	"os"
	"strings"
)

type Queue struct {
	SockWr int `json:"swr"` // per frames about 1572(1514+4+20+20+14)bytes
	SockRd int `json:"srd"` // per frames
	TapWr  int `json:"twr"` // per frames about 1572((1514+4+20+20+14))bytes
	TapRd  int `json:"trd"` // per frames
}

var (
	QdSwr = 1024 * 4
	QdSrd = 8
	QdTwr = 1024 * 2
	QdTrd = 2
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
	return libol.GenToken(13)
}

func GetTlsCfg(cfg Cert) *tls.Config {
	if cfg.KeyFile != "" && cfg.CrtFile != "" {
		cer, err := tls.LoadX509KeyPair(cfg.CrtFile, cfg.KeyFile)
		if err != nil {
			libol.Error("NewSwitch: %s", err)
		}
		return &tls.Config{Certificates: []tls.Certificate{cer}}
	}
	return nil
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
