package libol

import (
    "errors"
    "encoding/binary"
)

const (
    ETH_P_VLAN  = 0x8100
)

type Vlan struct {
    Tci uint8
    Vid uint16
    Pro uint16
}

func NewVlan(tci uint8, vid uint16) (this *Vlan) {
    this = &Vlan {
        Tci: tci,
        Vid: vid,
    }

    return
}

func (this *Vlan) Decode(frame []byte) error {
    if len(frame) < 4 {
        return errors.New("too small header") 
    }

    v := binary.BigEndian.Uint16(frame[0:2])
    this.Tci = uint8(0x000f & (v >> 12))
    this.Vid = uint16(0x0fff & v)
    this.Pro = binary.BigEndian.Uint16(frame[2:4]) 

    return nil
}

func (this *Vlan) Encode()[]byte {
    buffer := make([]byte, 16)

    v := (uint16(this.Tci) << 12) | this.Vid

    binary.BigEndian.PutUint16(buffer[0:2], v)
    binary.BigEndian.PutUint16(buffer[2:4], this.Pro)

    return buffer[:4]
}
