package main

import (
    "log"
    "flag"
    "fmt"

    "github.com/songgao/water"
    "github.com/danieldin95/openlan-go/olv1/olv1"
    "github.com/danieldin95/openlan-go/olv1/cpe"
)

type Cpe struct {
    verbose int
    tcpwroker *olv1cpe.TcpWroker 
    tapwroker *olv1cpe.TapWroker
}

func NewCpe(client *olv1.TcpClient, ifce *water.Interface, ifmtu int, verbose int) (this *Cpe){
    this = &Cpe {
        verbose: verbose,
        tapwroker : olv1cpe.NewTapWoker(ifce, ifmtu, verbose),
        tcpwroker : olv1cpe.NewTcpWoker(client, ifmtu, verbose),
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
    addr := flag.String("addr", "openlan.net:10002",  "the server connect to")
    verbose := flag.Int("verbose", 0x00, "open verbose")
    ifmtu := flag.Int("ifmtu", 1514, "the interface MTU include ethernet")

    flag.Parse()

    ifce := NewIfce(water.TAP)
    client := olv1.NewTcpClient(*addr, *verbose)
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
