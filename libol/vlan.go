package libol

import (
	"encoding/binary"
)

type Vlan struct {
	Tci uint16
	Vid uint16
	Pro uint16
	Len int
}

func NewVlan(tci uint16, vid uint16) (n *Vlan) {
	n = &Vlan{
		Tci: tci,
		Vid: vid,
		Len: 4,
	}

	return
}

func NewVlanFromFrame(frame []byte) (n *Vlan, err error) {
	n = &Vlan{
		Len: 4,
	}
	err = n.Decode(frame)
	return
}

func (n *Vlan) Decode(frame []byte) error {
	if len(frame) < 4 {
		return NewErr("Vlan.Decode: too small header")
	}

	v := binary.BigEndian.Uint16(frame[0:2])
	n.Tci = uint16(v >> 12)
	n.Vid = uint16(0x0fff & v)
	n.Pro = binary.BigEndian.Uint16(frame[2:4])

	return nil
}

func (n *Vlan) Encode() []byte {
	buffer := make([]byte, 16)

	v := (n.Tci << 12) | n.Vid
	binary.BigEndian.PutUint16(buffer[0:2], v)
	binary.BigEndian.PutUint16(buffer[2:4], n.Pro)

	return buffer[:4]
}
