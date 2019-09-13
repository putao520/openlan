package openlanv1

import (
    "encoding/binary"
)

type Arp struct {
    HrdCode uint16 // format hardware address
    ProCode uint16 // format protocol address
    HrdLen  uint8  // length of hardware address
    ProLen  uint8  // length of protocol address
    OpCode  uint16 // ARP Op(command)

    SHwAddr byte[] // sender hardware address.
    SIpAddr byte[] // sender IP address.
    THwAddr byte[] // target hardware address.
    TIpAddr byte[] // target IP address.
}

func NewArp() (this *Arp) {
    this = &Arp {
    }

    return
}

func NewArpFromFrame(frame byte[]) (this *Arp) {
    this = &Arp {
    }

    
    return
}