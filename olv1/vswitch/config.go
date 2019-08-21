package vswitch

import (
    "flag"
)

type Config struct {
    Brname string
    Verbose int
    HttpListen string
    TcpListen string
    Ifmtu int
    Token string
    TokenFile string
    Password string
} 

func NewConfig() (this *Config) {
    this = &Config {}

    flag.StringVar(&this.Brname, "br", "",  "the bridge name")
    flag.IntVar(&this.Verbose, "verbose", 0x00, "open verbose")
    flag.StringVar(&this.HttpListen, "http", "0.0.0.0:10082",  "the http listen on")
    flag.StringVar(&this.TcpListen, "addr", "0.0.0.0:10002",  "the server listen on")
    flag.IntVar(&this.Ifmtu, "ifmtu", 1514, "the interface MTU include ethernet")
    flag.StringVar(&this.Token, "token", "", "Administrator token")
    flag.StringVar(&this.TokenFile, "tokenfile", ".vswitch_oken", "The file administrator token saved to.")
    flag.StringVar(&this.Password, "password", ".password", "The file password loading from.")

    flag.Parse()

    return
}
