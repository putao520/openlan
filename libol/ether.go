package libol

import (
	"encoding/binary"
)

var (
	ZEROETHADDR    = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	BROADETHADDR   = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DEFAULTETHADDR = []byte{0x00, 0x16, 0x3e, 0x02, 0x56, 0x23}
)

const (
	ETH_P_ARP  = 0x0806
	ETH_P_IP4  = 0x0800
	ETH_P_IP6  = 0x86DD
	ETH_P_VLAN = 0x8100
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
		Src:  ZEROETHADDR,
		Dst:  ZEROETHADDR,
		Len:  14,
	}
	return
}

func NewEtherArp() (e *Ether) {
	return NewEther(ETH_P_ARP)
}

func NewEtherIP4() (e *Ether) {
	return NewEther(ETH_P_IP4)
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

	copy(buffer[:6], e.Dst)
	copy(buffer[6:12], e.Src)
	binary.BigEndian.PutUint16(buffer[12:14], e.Type)

	return buffer[:14]
}

func (e *Ether) IsVlan() bool {
	return e.Type == ETH_P_VLAN
}

func (e *Ether) IsArp() bool {
	return e.Type == ETH_P_ARP
}

func (e *Ether) IsIP4() bool {
	return e.Type == ETH_P_IP4
}
