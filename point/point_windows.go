package point

import (
    "net"
    
    "github.com/songgao/water"
    "github.com/milosgajdos83/tenus"
    "github.com/lightstar-dev/openlan-go/libol"
)

type PointWin struct {
    Client *libol.TcpClient
    Ifce *water.Interface
    Brname string
    Ifaddr string
    Ifname string
    
    //
    tcpwroker *TcpWroker 
    tapwroker *TapWroker
    br tenus.Bridger
    brip net.IP
    brnet *net.IPNet
    verbose int
}

func NewPointWin(config *Config) (this *PointWin) {
    ifce, err := water.New(water.Config { DeviceType: water.TAP })
    if err != nil {
        libol.Fatal("NewPointWin: ", err)
    }

    libol.Info("NewPointWin.device %s\n", ifce.Name())

    client := libol.NewTcpClient(config.Addr, config.Verbose)

    this = &PointWin {
        verbose: config.Verbose,
        Client: client,
        Ifce: ifce,
        Brname: config.Brname,
        Ifaddr: config.Ifaddr,
        Ifname: ifce.Name(),
        tapwroker : NewTapWoker(ifce, config),
        tcpwroker : NewTcpWoker(client, config),
    }
    return 
}

func (this *PointWin) Start() {
    if this.IsVerbose() {
        libol.Debug("PointWin.Start linux.\n")
    }

    if err := this.Client.Connect(); err != nil {
        libol.Error("PointWin.Start %s\n", err)
    }

    go this.tapwroker.GoRecv(this.tcpwroker.DoSend)
    go this.tapwroker.GoLoop()

    go this.tcpwroker.GoRecv(this.tapwroker.DoSend)
    go this.tcpwroker.GoLoop()
}

func (this *PointWin) Close() {
    this.Client.Close()
    this.Ifce.Close()

    if this.br != nil && this.brip != nil {
        if err := this.br.UnsetLinkIp(this.brip, this.brnet); err != nil {
            libol.Error("PointWin.Close.UnsetLinkIp %s : %s", this.br.NetInterface().Name, err)
        }
    }    
}

func (this *PointWin) UpLink() error {
    name := this.Ifce.Name()
    
    libol.Debug("PointWin.UpLink: %s", name)
    link, err := tenus.NewLinkFrom(name)
    if err != nil {
        libol.Error("PointWin.UpLink: Get ifce %s: %s", name, err)
        return err
    }
    
    if err := link.SetLinkUp(); err != nil {
        libol.Error("PointWin.UpLink.SetLinkUp: %s : %s", name, err)
        return err
    }

    return nil
}

func (this *PointWin) IsVerbose() bool {
    return this.verbose != 0
}