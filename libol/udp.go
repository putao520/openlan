package libol

import "encoding/binary"

const UDP_LEN = 8

type Udp struct {
	Source      uint16
	Destination uint16
	Length      uint16
	Checksum    uint16
	Len         int
}

func NewUdp() (u *Udp) {
	u = &Udp{
		Source:      0,
		Destination: 0,
		Length:      0,
		Checksum:    0,
		Len:         UDP_LEN,
	}
	return
}

func NewUdpFromFrame(frame []byte) (u *Udp, err error) {
	u = NewUdp()
	err = u.Decode(frame)
	return
}

func (u *Udp) Decode(frame []byte) error {
	if len(frame) < UDP_LEN {
		return NewErr("Udp.Decode: too small header: %d", len(frame))
	}

	u.Source = binary.BigEndian.Uint16(frame[0:2])
	u.Destination = binary.BigEndian.Uint16(frame[2:4])
	u.Length = binary.BigEndian.Uint16(frame[4:6])
	u.Checksum = binary.BigEndian.Uint16(frame[6:8])

	return nil
}

func (u *Udp) Encode() []byte {
	buffer := make([]byte, 32)

	binary.BigEndian.PutUint16(buffer[0:2], u.Source)
	binary.BigEndian.PutUint16(buffer[2:4], u.Destination)
	binary.BigEndian.PutUint16(buffer[4:6], u.Length)
	binary.BigEndian.PutUint16(buffer[6:8], u.Checksum)

	return buffer[:u.Len]
}
