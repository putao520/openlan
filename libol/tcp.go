package libol

import (
	"bytes"
	"encoding/binary"
)

const TCP_LEN = 20

type Tcp struct {
	Source         uint16
	Destination    uint16
	Sequence       uint32
	Acknowledgment uint32
	DataOffset     uint8
	ControlBits    uint8
	Window         uint16
	Checksum       uint16
	UrgentPointer  uint16
	Options        []byte
	Padding        []byte
	Len            int
}

func NewTcp() (t *Tcp) {
	t = &Tcp{
		Source:         0,
		Destination:    0,
		Sequence:       0,
		Acknowledgment: 0,
		DataOffset:     0,
		ControlBits:    0,
		Window:         0,
		Checksum:       0,
		UrgentPointer:  0,
		Len:            TCP_LEN,
	}
	return
}

func NewTcpFromFrame(frame []byte) (t *Tcp, err error) {
	t = NewTcp()
	err = t.Decode(frame)
	return
}

func (t *Tcp) Decode(frame []byte) error {
	if len(frame) < TCP_LEN {
		return Errer("Tcp.Decode: too small header: %d", len(frame))
	}

	reader := bytes.NewReader(frame)
	_ = binary.Read(reader, binary.BigEndian, &t.Source)
	_ = binary.Read(reader, binary.BigEndian, &t.Destination)
	_ = binary.Read(reader, binary.BigEndian, &t.Sequence)
	_ = binary.Read(reader, binary.BigEndian, &t.Acknowledgment)
	_ = binary.Read(reader, binary.BigEndian, &t.DataOffset)
	_ = binary.Read(reader, binary.BigEndian, &t.ControlBits)
	_ = binary.Read(reader, binary.BigEndian, &t.Window)
	_ = binary.Read(reader, binary.BigEndian, &t.Checksum)
	_ = binary.Read(reader, binary.BigEndian, &t.UrgentPointer)

	return nil
}

func (t *Tcp) Encode() []byte {
	writer := new(bytes.Buffer)

	_ = binary.Write(writer, binary.BigEndian, &t.Source)
	_ = binary.Write(writer, binary.BigEndian, &t.Destination)
	_ = binary.Write(writer, binary.BigEndian, &t.Sequence)
	_ = binary.Write(writer, binary.BigEndian, &t.Acknowledgment)
	_ = binary.Write(writer, binary.BigEndian, &t.DataOffset)
	_ = binary.Write(writer, binary.BigEndian, &t.ControlBits)
	_ = binary.Write(writer, binary.BigEndian, &t.Window)
	_ = binary.Write(writer, binary.BigEndian, &t.Checksum)
	_ = binary.Write(writer, binary.BigEndian, &t.UrgentPointer)

	return writer.Bytes()
}
