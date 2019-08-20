package controller

import (
    "flag"
)

type Config struct {
    Verbose bool
    UdpListen string
    Ifmtu int
    HttpListen string
    Token string
    TokenFile string
} 

func NewConfig() (this *Config) {
    this = &Config {}

    flag.BoolVar(&this.Verbose, "verbose", false, "open verbose")
    flag.StringVar(&this.UdpListen, "udp", "0.0.0.0:10020",  "the udp listen on")
    flag.IntVar(&this.Ifmtu, "ifmtu", 1430, "the interface MTU include ethernet") //1500-20-8-20-14 ~ 1430 for windows.
    flag.StringVar(&this.Token, "token", "", "the token for http authentication")
    flag.StringVar(&this.TokenFile, "tokenfile", ".controller_token", "the file token saved to")
    flag.StringVar(&this.HttpListen, "http", "0.0.0.0:10088",  "the http listen on")

    flag.Parse()

    return
}