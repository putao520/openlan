package libol

import (
    "errors"
    "encoding/binary"
)

var (
    ZEROETHADDR = []byte {0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
    BROADETHADDR = []byte {0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
    DEFAULTETHADDR = []byte {0x00, 0x16, 0x3e, 0x02, 0x56, 0x23}
)

const (
    ETH_P_ARP  = 0x0806
    ETH_P_IP4  = 0x0800
    ETH_P_IP6  = 0x86DD
    ETH_P_VLAN = 0x8100
)

type Ether struct {
    Dst []byte
    Src []byte
    Type uint16
    Len int
}

func NewEther(t uint16) (this *Ether) {
    this = &Ether {
        Type: t,
        Src: ZEROETHADDR,
        Dst: ZEROETHADDR,
        Len: 14,
    }
    return
}

func NewEtherFromFrame(frame []byte) (this *Ether, err error) {
    this = &Ether {
        Len: 14,
    }
    err = this.Decode(frame)
    return
}

func (this *Ether) Decode(frame []byte) error {
    if len(frame) < 14 {
        return errors.New("too small header") 
    }

    this.Dst = frame[:6]
    this.Src = frame[6:12]
    this.Type = binary.BigEndian.Uint16(frame[12:14])
    this.Len = 14

    return nil
}

func (this *Ether) Encode() []byte {
    buffer := make([]byte, 14)

    copy(buffer[:6], this.Dst)
    copy(buffer[6:12], this.Src)
    binary.BigEndian.PutUint16(buffer[12:14], this.Type)

    return buffer[:14]
}
