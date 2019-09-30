package libol

import (
    "encoding/binary"
    "errors"
)

type Vlan struct {
    Tci uint8
    Vid uint16
    Pro uint16
    Len int
}

func NewVlan(tci uint8, vid uint16) (this *Vlan) {
    this = &Vlan {
        Tci: tci,
        Vid: vid,
        Len: 4,
    }

    return
}

func NewVlanFromFrame(frame []byte) (this *Vlan, err error) {
    this = &Vlan {
        Len: 4,
    }
    err = this.Decode(frame)
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

func (this *Vlan) Encode() []byte {
    buffer := make([]byte, 16)

    v := (uint16(this.Tci) << 12) | this.Vid

    binary.BigEndian.PutUint16(buffer[0:2], v)
    binary.BigEndian.PutUint16(buffer[2:4], this.Pro)

    return buffer[:4]
}
