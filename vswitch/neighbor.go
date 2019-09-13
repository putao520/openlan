package vswitch

type Neighbor struct {
    HwAddr string `json:"HwAddr"`
    IpAddr string `json:"IpAddr"`
}

func NewNeighbor(hwaddr string, ipaddr string) (this *Neighbor) {
    this = &Neighbor {
        HwAddr: hwaddr,
        IpAddr: ipaddr,
    }

    return
}
