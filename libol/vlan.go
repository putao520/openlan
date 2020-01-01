package libol

import (
	"bytes"
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
	var err error

	if len(frame) < 4 {
		return Errer("Vlan.Decode: too small header")
	}

	reader := bytes.NewReader(frame)

	v := uint16(0)
	err = binary.Read(reader, binary.BigEndian, &v)
	n.Tci = 0x0f & (v >> 12)
	n.Vid = 0x0fff & v
	err = binary.Read(reader, binary.BigEndian, &n.Pro)

	return err
}

func (n *Vlan) Encode() []byte {
	writer := new(bytes.Buffer)

	v := (n.Tci << 12) | n.Vid
	_ = binary.Write(writer, binary.BigEndian, &v)
	_ = binary.Write(writer, binary.BigEndian, &n.Pro)

	return writer.Bytes()
}
