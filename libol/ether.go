package libol

import (
	"encoding/binary"
)

var (
	ZEROED    = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	BROADED   = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DEFAULTED = []byte{0x00, 0x16, 0x3e, 0x02, 0x56, 0x23}
)

const (
	ETHPARP  = 0x0806
	ETHPIP4  = 0x0800
	ETHPIP6  = 0x86DD
	ETHPVLAN = 0x8100
)

type Ether struct {
	Dst  []byte
	Src  []byte
	Type uint16
	Len  int
}

func NewEther(t uint16) (e *Ether) {
	e = &Ether{
		Type: t,
		Src:  ZEROED,
		Dst:  ZEROED,
		Len:  14,
	}
	return
}

func NewEtherArp() (e *Ether) {
	return NewEther(ETHPARP)
}

func NewEtherIP4() (e *Ether) {
	return NewEther(ETHPIP4)
}

func NewEtherFromFrame(frame []byte) (e *Ether, err error) {
	e = &Ether{
		Len: 14,
	}
	err = e.Decode(frame)
	return
}

func (e *Ether) Decode(frame []byte) error {
	if len(frame) < 14 {
		return Errer("Ether.Decode too small header: %d", len(frame))
	}

	e.Dst = frame[:6]
	e.Src = frame[6:12]
	e.Type = binary.BigEndian.Uint16(frame[12:14])
	e.Len = 14

	return nil
}

func (e *Ether) Encode() []byte {
	buffer := make([]byte, 14)

<<<<<<< HEAD
	_ = binary.Write(writer, binary.BigEndian, e.Dst[:6])
	_ = binary.Write(writer, binary.BigEndian, e.Src[:6])
	_ = binary.Write(writer, binary.BigEndian, &e.Type)
=======
	copy(buffer[:6], e.Dst)
	copy(buffer[6:12], e.Src)
	binary.BigEndian.PutUint16(buffer[12:14], e.Type)
>>>>>>> parent of e188bfd... arrange decode/encode on vlan.

	return buffer[:14]
}

func (e *Ether) IsVlan() bool {
	return e.Type == ETHPVLAN
}

func (e *Ether) IsArp() bool {
	return e.Type == ETHPARP
}

func (e *Ether) IsIP4() bool {
	return e.Type == ETHPIP4
}
