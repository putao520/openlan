package point

import (
    "github.com/songgao/water"
    "github.com/lightstar-dev/openlan-go/libol"
)

type Point struct {
    Verbose int
    Client *libol.TcpClient
    Ifce *water.Interface
    //
    tcpwroker *TcpWroker 
    tapwroker *TapWroker
}

func NewPoint(config *Config) (this *Point){
    ifce, err := water.New(water.Config { DeviceType: water.TAP })
    if err != nil {
        libol.Fatal("NewPoint: ", err)
    }
    libol.Info("NewPoint.device %s\n", ifce.Name())

    client := libol.NewTcpClient(config.Addr, config.Verbose)

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
        libol.Error("Point.Start %s\n", err)
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
