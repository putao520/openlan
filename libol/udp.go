package libol

import (
	"bytes"
	"encoding/binary"
)

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
	var err error

	if len(frame) < UDP_LEN {
		return Errer("Udp.Decode: too small header: %d", len(frame))
	}

	reader := bytes.NewReader(frame)

	err = binary.Read(reader, binary.BigEndian, &u.Source)
	err = binary.Read(reader, binary.BigEndian, &u.Destination)
	err = binary.Read(reader, binary.BigEndian, &u.Length)
	err = binary.Read(reader, binary.BigEndian, &u.Checksum)

	return err
}

func (u *Udp) Encode() []byte {
	writer := new(bytes.Buffer)

	_ = binary.Write(writer, binary.BigEndian, &u.Source)
	_ = binary.Write(writer, binary.BigEndian, &u.Destination)
	_ = binary.Write(writer, binary.BigEndian, &u.Length)
	_ = binary.Write(writer, binary.BigEndian, &u.Checksum)

	return writer.Bytes()
}
