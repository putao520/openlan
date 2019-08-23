package point

import (
    "flag"
    "strings"
    "fmt"
)

type Config struct {
    Addr string `json:"vsAddr"`
    Verbose int `json:"verbose"`
    Ifmtu int `json:"ifMtu"`
    Auth string `json:"vsAuth"`
    
    Name string 
    Password string 
}

func RightAddr(listen *string, port int) {
    values := strings.Split(*listen, ":")
    if len(values) == 1 {
        *listen = fmt.Sprintf("%s:%d", values[0], port)
    }
}

func NewConfig() (this *Config) {
    this = &Config {}

    flag.StringVar(&this.Addr, "vs:addr", "openlan.net:10002",  "the server connect to")
    flag.IntVar(&this.Verbose, "verbose", 0x00, "open verbose")
    flag.IntVar(&this.Ifmtu, "if:mtu", 1514, "the interface MTU include ethernet")
    flag.StringVar(&this.Auth, "vs:auth", "openlan:password",  "the auth login to")
    
    flag.Parse()
    
    values := strings.Split(this.Auth, ":")
    this.Name = values[0] 
    if (len(values) > 1) {
        this.Password = values[1]
    }

    RightAddr(&this.Addr, 10002)
    
    return
}