package endpoint

import (
	"flag"
	"strings"
	"fmt"
)

type Config struct {
	Verbose bool
	UdpListen string
	Controller string
	HttpListen string
	Ifmtu int
	Auth string
	Interval int
	Token string
	TokenFile string
}

func RightListen(listen *string, port int) {
	values := strings.Split(*listen, ":")
	if len(values) == 1 {
		*listen = fmt.Sprintf("%s:%d", values[0], port)
	}
}

func NewConfig() (this *Config) {
	this = &Config {}

    flag.BoolVar(&this.Verbose, "verbose", false, "open verbose")
    flag.StringVar(&this.UdpListen, "udp", "0.0.0.0:10010",  "the udp listen on")
	flag.StringVar(&this.Controller, "ctl", "openlan.net:10020",  "the controller listen on")
	flag.StringVar(&this.HttpListen, "http", "0.0.0.0:10082",  "the http listen on")
	flag.StringVar(&this.Auth, "auth", "default@openlan:", "the authentication login")
	flag.IntVar(&this.Ifmtu, "ifmtu", 1438, "the interface MTU include ethernet") //1500-20-8-20-14
	flag.IntVar(&this.Interval, "interval", 30, "the interval heartbeat to controller") 
	flag.StringVar(&this.Token, "token", "", "the token for http authentication")
	flag.StringVar(&this.TokenFile, "tokenfile", ".endpoint_token", "the file token saved to")

	flag.Parse()

	RightListen(&this.Controller, 10020)
	RightListen(&this.UdpListen,  10010)
	RightListen(&this.HttpListen, 10082)

	return
}