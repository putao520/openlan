package libol

import (
    "encoding/binary"
    "errors"
)

const (
    ARP_REQUEST = 1
    ARP_REPLY   = 2
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

    SHwAddr []byte // sender hardware address.
    SIpAddr []byte // sender IP address.
    THwAddr []byte // target hardware address.
    TIpAddr []byte // target IP address.
    Len int
}

func NewArp() (this *Arp) {
    this = &Arp {
        HrdCode: ARPHRD_ETHER,
        ProCode: ETH_P_IP4,
        HrdLen: 6,
        ProLen: 4,
        OpCode: ARP_REQUEST,
        Len: 0,
    }

    return
}

func NewArpFromFrame(frame []byte) (this *Arp, err error) {
    this = &Arp {
        Len: 0,
    }
    err = this.Decode(frame)
    return
}

func (this *Arp) Decode(frame []byte) error {
    if len(frame) < 8 {
        return errors.New("too small header") 
    }

    this.HrdCode = binary.BigEndian.Uint16(frame[0:2])
    this.ProCode = binary.BigEndian.Uint16(frame[2:4])
    this.HrdLen  = uint8(frame[4])
    this.ProLen  = uint8(frame[5])
    this.OpCode  = binary.BigEndian.Uint16(frame[6:8])

    p := uint8(8)
    if len(frame) < int(p + 2 * (this.HrdLen + this.ProLen)) {
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

    this.Len = int(p)

    return nil
}

func (this *Arp) Encode() []byte {
    buffer := make([]byte, 1024)
    
    binary.BigEndian.PutUint16(buffer[0:2], this.HrdCode)
    binary.BigEndian.PutUint16(buffer[2:4], this.ProCode)
    buffer[4] = byte(this.HrdLen)
    buffer[5] = byte(this.ProLen)
    binary.BigEndian.PutUint16(buffer[6:8], this.OpCode)

    p := uint8(8)
    copy(buffer[p:p+this.HrdLen], this.SHwAddr[0:this.HrdLen])
    p += this.HrdLen
    copy(buffer[p:p+this.ProLen], this.SIpAddr[0:this.ProLen])
    p += this.ProLen

    copy(buffer[p:p+this.HrdLen], this.THwAddr[0:this.HrdLen])
    p += this.HrdLen
    copy(buffer[p:p+this.ProLen], this.TIpAddr[0:this.ProLen])
    p += this.ProLen

    this.Len = int(p)

    return buffer[:p]
}