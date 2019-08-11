package main

import (
    "log"
    "flag"
    "fmt"

    "./openlan"
    "github.com/songgao/water"
)

type Cpe struct {
    verbose int
    tcpwroker *openlan.TcpWroker 
    tapwroker *openlan.TapWroker
}

func NewCpe(client *openlan.TcpClient, ifce *water.Interface, ifmtu int, verbose int) (this *Cpe){
    this = &Cpe {
        verbose: verbose,
        tapwroker : openlan.NewTapWoker(ifce, ifmtu, verbose),
        tcpwroker : openlan.NewTcpWoker(client, ifmtu, verbose),
    }
    return 
}

func (this *Cpe) Start() {
    go this.tapwroker.GoRecv(this.tcpwroker.DoSend)
    go this.tapwroker.GoLoop()

    go this.tcpwroker.GoRecv(this.tapwroker.DoSend)
    go this.tcpwroker.GoLoop()
}

func NewIfce(devtype water.DeviceType) (ifce *water.Interface) {
    ifce, err := water.New(water.Config {
        DeviceType: devtype,
    })
    if err != nil {
        log.Fatal(err)
    }

    return 
}

func main() {
    addr := flag.String("addr", "openlan.net:10001",  "the server address")
    verbose := flag.Int("verbose", 0x00, "open verbose")
    ifmtu := flag.Int("ifmtu", 1514, "the interface MTU include ethernet")

    flag.Parse()

    ifce := NewIfce(water.TAP)
    client := openlan.NewTcpClient(*addr, *verbose)
    cpe := NewCpe(client, ifce, *ifmtu, *verbose)
    
    cpe.Start()

    var input string

    for {
        fmt.Println("Please press enter `q` to exit...")
        fmt.Scanln(&input)
        if input == "q" {
            break
        }
    }
    fmt.Println("Done!")
}
