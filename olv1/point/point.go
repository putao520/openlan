package point

import (
    "log"

    "github.com/songgao/water"
    "github.com/danieldin95/openlan-go/olv1/openlanv1"
)

type Point struct {
    Verbose int
    Client *openlanv1.TcpClient
    Ifce *water.Interface
    //
    tcpwroker *TcpWroker 
    tapwroker *TapWroker
}

func NewPoint(config *Config) (this *Point){
    ifce, err := water.New(water.Config { DeviceType: water.TAP })
    if err != nil {
        log.Fatal(err)
    }
    
    client := openlanv1.NewTcpClient(config.Addr, config.Verbose)

    this = &Point {
        Verbose: config.Verbose,
        Client: client,
        Ifce: ifce,
        tapwroker : NewTapWoker(ifce, config),
        tcpwroker : NewTcpWoker(client, config),
    }
    return 
}

func (this *Point) Start() {
    if err := this.Client.Connect(); err != nil {
        log.Printf("Error| Point.Start %s\n", err)
    }

    go this.tapwroker.GoRecv(this.tcpwroker.DoSend)
    go this.tapwroker.GoLoop()

    go this.tcpwroker.GoRecv(this.tapwroker.DoSend)
    go this.tcpwroker.GoLoop()
}

func (this *Point) Close() {
    this.Client.Close()
    this.Ifce.Close()
}