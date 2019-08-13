package main

import (
    "log"
    "flag"
    "fmt"

    "github.com/songgao/water"
    "github.com/danieldin95/openlan-go-v1"
)

type Cpe struct {
    verbose int
    tcpwroker *openlanv1.TcpWroker 
    tapwroker *openlanv1.TapWroker
}

func NewCpe(client *openlanv1.TcpClient, ifce *water.Interface, ifmtu int, verbose int) (this *Cpe){
    this = &Cpe {
        verbose: verbose,
        tapwroker : openlanv1.NewTapWoker(ifce, ifmtu, verbose),
        tcpwroker : openlanv1.NewTcpWoker(client, ifmtu, verbose),
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
    client := openlanv1.NewTcpClient(*addr, *verbose)
    cpe := NewCpe(client, ifce, *ifmtu, *verbose)
    
    cpe.Start()

    for {
        var input string

        fmt.Println("Please press enter `q` to exit...")
        if fmt.Scanln(&input); input == "q" {
            break
        }
    }

    client.Close()
    fmt.Println("Done!")
}
