package vswitch

import (
    "fmt"
    "net"
    "log"
    "github.com/danieldin95/openlan-go/libol"
)

type Neighbor struct {
    Client *libol.TcpClient  `json:"Client"`

    HwAddr net.HardwareAddr `json:"HwAddr"`
    IpAddr net.IP `json:"IpAddr"`
}

func (this *Neighbor) String() string {
    return fmt.Sprintf("%s,%s,%s", this.HwAddr, this.IpAddr, this.Client)
}

func NewNeighbor(hwaddr net.HardwareAddr, ipaddr net.IP, client *libol.TcpClient) (this *Neighbor) {
    this = &Neighbor {
        HwAddr: hwaddr,
        IpAddr: ipaddr,
        Client: client,
    }

    return
}

type Neighborer struct {
    Neighbors map[string]*Neighbor
    verbose int
}

func NewNeighborer(c *Config) (this *Neighborer) {
    this = &Neighborer {
        Neighbors: make(map[string]*Neighbor, 1024*10),
        verbose: c.Verbose,
    }

    return
}

func (this *Neighborer) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
    if this.IsVerbose() {
        log.Printf("Debug| Neighborer.OnFrame % x.", frame.Data)
    }

    if libol.IsInst(frame.Data) {
        return nil
    }

    ethtype := frame.EthType()
    ethdata := frame.EthData()
    if ethtype != libol.ETH_P_ARP {
        if ethtype == libol.ETH_P_VLAN {
            //TODO
        }
        
        return nil
    }

    arp, err := libol.NewArpFromFrame(ethdata)
    if err != nil {
        log.Printf("Error| Neighborer.OnFrame %s.", err)
        return nil
    }

    if arp.ProCode == libol.ETH_P_IP4 {
        if arp.OpCode == libol.ARP_REQUEST ||
           arp.OpCode == libol.ARP_REPLY {
            n := NewNeighbor(net.HardwareAddr(arp.SHwAddr), net.IP(arp.SIpAddr), client)
            this.AddNeighbor(n)
        }
    }
    
    return nil
}

func (this *Neighborer) AddNeighbor(neb *Neighbor) {
    if n := this.Neighbors[neb.HwAddr.String()]; n != nil {
        //TODO update.
        log.Printf("Info| Neighborer.AddNeighbor: update %s.", neb)
    } else {
        log.Printf("Info| Neighborer.AddNeighbor: new %s.", neb)
    }
    
    this.Neighbors[neb.HwAddr.String()] = neb

    //TODO publish via redis.
}

func (this *Neighborer) DelNeighbor(hwaddr net.HardwareAddr) {
    log.Printf("Info| Neighborer.DelNeighbor %s.", hwaddr)
    if n := this.Neighbors[hwaddr.String()]; n != nil {
        delete(this.Neighbors, hwaddr.String())
    }
}

func (this *Neighborer) OnClientClose(client *libol.TcpClient) {
    //TODO
    log.Printf("Info| Neighborer.OnClientClose %s.", client)
}

func (this *Neighborer) IsVerbose() bool {
    return this.verbose != 0
}
