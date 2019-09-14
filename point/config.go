package point

import (
    "flag"
    "strings"
    "fmt"
)

type Config struct {
    Addr string `json:"vsAddr"`
    Auth string `json:"vsAuth"`
    Verbose int `json:"verbose"`
    Ifmtu int `json:"ifMtu"`
    Ifaddr string `json:"ifAddr"`
    
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
    flag.StringVar(&this.Auth, "vs:auth", "openlan:password",  "the auth login to")
    flag.IntVar(&this.Verbose, "verbose", 0x00, "open verbose")
    flag.IntVar(&this.Ifmtu, "if:mtu", 1518, "the interface MTU include ethernet")
    flag.StringVar(&this.Ifaddr, "if:addr", "192.168.1.254/24", "the interface address")
    
    flag.Parse()
    
    values := strings.Split(this.Auth, ":")
    this.Name = values[0] 
    if (len(values) > 1) {
        this.Password = values[1]
    }

    RightAddr(&this.Addr, 10002)
    
    return
}