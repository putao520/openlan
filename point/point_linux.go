package point

import (
    "net"
    
    "github.com/songgao/water"
    "github.com/milosgajdos83/tenus"
    "github.com/lightstar-dev/openlan-go/libol"
)

type Point struct {
    Client *libol.TcpClient
    Ifce   *water.Interface
    Brname string
    Ifaddr string
    Ifname string

    tcpwroker *TcpWroker 
    tapwroker *TapWroker
    br        tenus.Bridger
    brip      net.IP
    brnet     *net.IPNet
}

func NewPoint(config *Config) (this *Point) {
    var err error
    var ifce *water.Interface

    if config.Iftun {
        ifce, err = water.New(water.Config { DeviceType: water.TUN })
    } else {
        ifce, err = water.New(water.Config { DeviceType: water.TAP })
    }
    if err != nil {
        libol.Fatal("NewPoint: ", err)
    }

    libol.Info("NewPoint.device %s\n", ifce.Name())
    client := libol.NewTcpClient(config.Addr)
    this = &Point {
        Client: client,
        Ifce  : ifce,
        Brname: config.Brname,
        Ifaddr: config.Ifaddr,
        Ifname: ifce.Name(),
        tapwroker : NewTapWoker(ifce, config),
        tcpwroker : NewTcpWoker(client, config),
    }
    return 
}

func (this *Point) Start() {
    libol.Debug("Point.Start linux.\n")

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

    if this.br != nil && this.brip != nil {
        if err := this.br.UnsetLinkIp(this.brip, this.brnet); err != nil {
            libol.Error("Point.Close.UnsetLinkIp %s: %s", this.br.NetInterface().Name, err)
        }
    }    
}

func (this *Point) UpLink() error {
    name := this.Ifce.Name()
    
    libol.Debug("Point.UpLink: %s", name)
    link, err := tenus.NewLinkFrom(name)
    if err != nil {
        libol.Error("Point.UpLink: Get ifce %s: %s", name, err)
        return err
    }
    
    if err := link.SetLinkUp(); err != nil {
        libol.Error("Point.UpLink.SetLinkUp: %s : %s", name, err)
        return err
    }

    if this.Brname != "" {
        br, err := tenus.BridgeFromName(this.Brname)
        if err != nil {
            libol.Error("Point.UpLink.newBr: %s", err)
            br, err = tenus.NewBridgeWithName(this.Brname)
            if err != nil {
                libol.Error("Point.UpLink.newBr: %s", err)
            }
        }
        if err := libol.BrCtlStp(this.Brname, true); err != nil {
            libol.Error("Point.UpLink.ctlstp: %s", err)
        }

        if err := br.SetLinkUp(); err != nil {
            libol.Error("Point.UpLink.newBr.Up: %s", err)
        }

        if err := br.AddSlaveIfc(link.NetInterface()); err != nil {
            libol.Error("Point.UpLink.AddSlave: Switch ifce %s: %s", name, err)
        }

        link, err = tenus.NewLinkFrom(this.Brname)
        if err != nil {
            libol.Error("Point.UpLink: Get ifce %s: %s", this.Brname, err)
        }

        this.br = br
    }

    if this.Ifaddr != "" {
        ip, ipnet, err := net.ParseCIDR(this.Ifaddr)
        if err != nil {
            libol.Error("Point.UpLink.ParseCIDR %s : %s", this.Ifaddr, err)
            return err
        }
        if err := link.SetLinkIp(ip, ipnet); err != nil {
            libol.Error("Point.UpLink.SetLinkIp : %s", err)
            return err
        }

        this.brip = ip
        this.brnet = ipnet
    }

    return nil
}