package endpoint

import (
	"flag"
)

type Config struct {
	Verbose bool
	UdpListen string
	Controller string
	Ifmtu int
	Auth string
} 

func NewConfig() (this *Config) {
	this = &Config {}

    flag.BoolVar(&this.Verbose, "verbose", false, "open verbose")
    flag.StringVar(&this.UdpListen, "udp", "0.0.0.0:10010",  "the udp listen on")
	flag.StringVar(&this.Controller, "ctl", "openlan.net:10020",  "the controller listen on")
	flag.StringVar(&this.Auth, "auth", "default@openlan:", "the authentication login")

	flag.Parse()

	return
}