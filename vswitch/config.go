package vswitch

import (
    "flag"
    "strings"
    "fmt"
)

type Config struct {
    Brname string `json:"vsBr"`
    Verbose int `json:"verbose"`
    HttpListen string `json:"httpAddr"`
    TcpListen string `json:"vsAddr"`
    Ifmtu int `json:"ifMtu"`
    Ifaddr string `json:"ifAddr"`
    Token string `json:"adminToken"`
    TokenFile string `json:"adminFile"`
    Password string `json:"authFile"`
} 

func RightAddr(listen *string, port int) {
    values := strings.Split(*listen, ":")
    if len(values) == 1 {
        *listen = fmt.Sprintf("%s:%d", values[0], port)
    }
}

func NewConfig() (this *Config) {
    this = &Config {}

    flag.StringVar(&this.Brname, "vs:br", "",  "the bridge name")
    flag.IntVar(&this.Verbose, "verbose", 0x00, "open verbose")
    flag.StringVar(&this.HttpListen, "http:addr", "0.0.0.0:10082",  "the http listen on")
    flag.StringVar(&this.TcpListen, "vs:addr", "0.0.0.0:10002",  "the server listen on")
    flag.StringVar(&this.Token, "admin:token", "", "Administrator token")
    flag.StringVar(&this.TokenFile, "admin:file", ".vswitch_oken", "The file administrator token saved to.")
    flag.StringVar(&this.Password, "auth:file", ".password", "The file password loading from.")
    flag.IntVar(&this.Ifmtu, "if:mtu", 1518, "the interface MTU include ethernet")
    flag.StringVar(&this.Ifaddr, "if:addr", "192.168.100.2/24", "the interface address")

    flag.Parse()
   
    RightAddr(&this.TcpListen, 10002)
    RightAddr(&this.HttpListen, 10082)

    return
}
