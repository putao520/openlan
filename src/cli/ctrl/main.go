package main

import (
	"flag"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olctl/http"
	"github.com/danieldin95/openlan-go/src/olctl/storage"
	"os"
)

type Config struct {
	StaticDir string `json:"dir.static"`
	CrtDir    string `json:"dir.crt"`
	ConfDir   string `json:"-"`
	Verbose   int    `json:"log.level"`
	LogFile   string `json:"log.file"`
	Listen    string `json:"listen"`
}

var cfg = Config{
	StaticDir: "dist",
	CrtDir:    "ca",
	ConfDir:   "/etc/openlan/ctrl",
	Listen:    "0.0.0.0:10088",
	LogFile:   "/var/log/openlan-ctrl.log",
	Verbose:   20,
}

func main() {
	flag.StringVar(&cfg.Listen, "listen", cfg.Listen, "the address http listen.")
	flag.IntVar(&cfg.Verbose, "log:level", cfg.Verbose, "logger level")
	flag.StringVar(&cfg.CrtDir, "crt:dir", cfg.CrtDir, "the directory X509 certificate file on.")
	flag.StringVar(&cfg.StaticDir, "dist:dir", cfg.StaticDir, "the dist directory.")
	flag.StringVar(&cfg.ConfDir, "conf", cfg.ConfDir, "the directory configuration on")
	flag.Parse()

	libol.PreNotify()
	libol.SetLogger(cfg.LogFile, cfg.Verbose)
	storage.Storager.Load(cfg.ConfDir)

	h := http.NewServer(cfg.Listen, cfg.StaticDir)
	if _, err := os.Stat(cfg.CrtDir); !os.IsNotExist(err) {
		h.SetCert(cfg.CrtDir+"/private.key", cfg.CrtDir+"/crt.pem")
	}
	go h.Start()
	libol.SdNotify()
	libol.Wait()
}
