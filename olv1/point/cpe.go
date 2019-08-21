package point

import (
    "log"

    "github.com/songgao/water"
    "github.com/danieldin95/openlan-go/olv1/olv1"
)

type Cpe struct {
    Verbose int
    Client *openlanv1.TcpClient
    Ifce *water.Interface
    //
    tcpwroker *TcpWroker 
    tapwroker *TapWroker
}

func NewCpe(config *Config) (this *Cpe){
    ifce, err := water.New(water.Config { DeviceType: water.TAP })
    if err != nil {
        log.Fatal(err)
    }
    
    client := openlanv1.NewTcpClient(config.Addr, config.Verbose)

    this = &Cpe {
        Verbose: config.Verbose,
        Client: client,
        Ifce: ifce,
        tapwroker : NewTapWoker(ifce, config),
        tcpwroker : NewTcpWoker(client, config),
    }
    return 
}

func (this *Cpe) Start() {
    if err := this.Client.Connect(); err != nil {
        log.Printf("Error| Cpe.Start %s\n", err)
    }

    go this.tapwroker.GoRecv(this.tcpwroker.DoSend)
    go this.tapwroker.GoLoop()

    go this.tcpwroker.GoRecv(this.tapwroker.DoSend)
    go this.tcpwroker.GoLoop()
}

func (this *Cpe) Close() {
    this.Client.Close()
    this.Ifce.Close()
}