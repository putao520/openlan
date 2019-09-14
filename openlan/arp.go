package openlanv1

import (
    "fmt"
    "errors"
    "encoding/binary"
)

const (
    ARP_REQUEST = 1
    ARP_REPLY   = 2
)

const (
    ETH_P_ARP  = 0x0806
    ETH_P_IP4  = 0x0800
    ETH_P_IP6  = 0x86DD
)

const (
    ARPHRD_NETROM = 0
    ARPHRD_ETHER  = 1
    ARPHRD_EETHER = 2
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

func (this *Arp) Decode(frame byte[]) error {
    if len(frame) < 8 {
        return errors.New("too small header") 
    }

    this.HrdCode = binary.BigEndian.Uint16(frame[0:2])
    this.ProCode = binary.BigEndian.Uint16(frame[2:4])
    this.HrdLen  = binary.BigEndian.Uint8(frame[4:5])
    this.ProLen  = binary.BigEndian.Uint8(frame[5:6])
    this.OpCode  = binary.BigEndian.Uint16(frame[6:8])

    p := 8
    if len(frame) < p + 2 * (this.HrdLen + this.ProLen) {
        return errors.New("too small frame") 
    }

    this.SHwAddr = frame[p:p+this.HrdLen]
    p += this.HrdLen
    this.SIpAddr = frame[p:p+this.ProLen]
    p += this.ProLen

    this.THwAddr = frame[p:p+this.HrdLen]
    p += this.HrdLen
    this.TIpAddr = frame[p:p+this.ProLen]
    p += this.ProLen

    return nil
}

func (this *Arp) Encode()[]byte {
    buffer := make([]byte, 1024)
    
    binary.BigEndian.PutUint16(buffer[0:2], this.HrdCode)
    binary.BigEndian.PutUint16(buffer[2:4], this.ProCode)
    binary.BigEndian.PutUint8(buffer[4:5], this.HrdLen)
    binary.BigEndian.PutUint8(buffer[5:6], this.ProLen)
    binary.BigEndian.PutUint16(buffer[6:8], this.OpCode)

    p := 8
    copy(buffer[p:p+this.HrdLen], this.SHwAddr[0:this.HrdLen])
    p += this.HrdLen
    copy(buffer[p:p+this.ProLen], this.SIpAddr[0:this.ProLen])
    p += this.ProLen

    copy(buffer[p:p+this.HrdLen], this.THwAddr[0:this.HrdLen])
    p += this.HrdLen
    copy(buffer[p:p+this.ProLen], this.TIpAddr[0:this.ProLen])
    p += this.ProLen

    return buffer[:p]
}