package controller

import (
	"flag"
)

type Config struct {
	Verbose bool
	UdpListen string
	Ifmtu int
} 

func NewConfig() (this *Config) {
	this = &Config {}

    flag.BoolVar(&this.Verbose, "verbose", false, "open verbose")
	flag.StringVar(&this.UdpListen, "udp", "0.0.0.0:10020",  "the udp listen on")
	flag.IntVar(&this.Ifmtu, "ifmtu", 1438, "the interface MTU include ethernet") //1500-20-8-20-14

	flag.Parse()

	return
}