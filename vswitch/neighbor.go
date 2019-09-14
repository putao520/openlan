package vswitch

import (
    "fmt"
    "net"
    "github.com/danieldin95/openlan-go/openlan"
)

type Neighbor struct {
    Client *openlan.TcpClient  `json:"Client"`

    HwAddr net.HardwareAddr `json:"HwAddr"`
    IpAddr net.IP `json:"IpAddr"`
}

func (this *Neighbor) String() {
    return fmt.Sprintf("%s,%s,%s", this.HwAddr, this.IpAddr, this.Client)
}

func NewNeighbor(hwaddr net.HardwareAddr, ipaddr net.IP, client *openlan.TcpClient) (this *Neighbor) {
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

func (this *Neighborer) OnFrame(client *openlan.TcpClient, frame *openlan.Frame) error {
    if this.IsVerbose() {
        log.Printf("Debug| Neighborer.OnFrame % x.", frame.Data)
    }

    if openlan.IsInst(frame.Data) {
        return nil
    }

    ethtype := frame.EthType()
    ethdata := frame.EthData()
    if ethtype != openlan.ETH_P_ARP {
        if ethtype == openlan.ETH_P_VLAN {
            //TODO
        }
        
        return nil
    }

    arp := openlan.NewArp()
    if err := arp.Decode(ethdata); err != nil {
        log.Printf("Error| Neighborer.OnFrame %s.", err)
        return nil
    }

    if arp.ProCode == openlan.ETH_P_IP4 {
        if arp.OpCode == openlan.ARP_REQUEST ||
           arp.OpCode == openlan.ARP_REPLY {
            n := NewNeighbor(net.HardwareAddr(arp.SHwAddr), net.IP(arp.SIpAddr), client)
            this.AddNeighbor(n)
        }
    }
    
    return nil
}

func (this *Neighborer) AddNeighbor(neb *Neighbor) {
    if this.IsVerbose() {
        log.Printf("Debug| Neighborer.AddNeighbor %s.", neb)
    }
    
    if n := this.Neighbors[neb.HwAddr.String()]; n != nil {
        //TODO update.
    }
    
    this.Neighbors[neb.HwAddr.String()] = neb

    //TODO publish via redis.
}

func (this *Neighborer) DelNeighbor(hwaddr net.HardwareAddr) {
    if n := this.Neighbors[hwaddr.String()]; n != nil {
        delete(this.Neighbors, hwaddr.String())
    }
}

func (this *Neighbor) OnClientClose(client *openlan.TcpClient) {
    //TODO
}

func (this *Neighbor) IsVerbose() bool {
    return this.verbose != 0
}
