package libol

import "encoding/binary"

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
		return NewErr("Tcp.Decode: too small header: %d", len(frame))
	}

	t.Source = binary.BigEndian.Uint16(frame[0:2])
	t.Destination = binary.BigEndian.Uint16(frame[2:4])
	t.Sequence = binary.BigEndian.Uint32(frame[4:8])
	t.Acknowledgment = binary.BigEndian.Uint32(frame[8:12])
	t.DataOffset = uint8(frame[12])
	t.ControlBits = uint8(frame[13])
	t.Window = binary.BigEndian.Uint16(frame[14:16])
	t.Checksum = binary.BigEndian.Uint16(frame[16:18])
	t.UrgentPointer = binary.BigEndian.Uint16(frame[18:20])

	return nil
}

func (t *Tcp) Encode() []byte {
	buffer := make([]byte, 32)

	binary.BigEndian.PutUint16(buffer[0:2], t.Source)
	binary.BigEndian.PutUint16(buffer[2:4], t.Destination)
	binary.BigEndian.PutUint32(buffer[4:8], t.Sequence)
	binary.BigEndian.PutUint32(buffer[8:12], t.Acknowledgment)
	buffer[12] = t.DataOffset
	buffer[13] = t.ControlBits
	binary.BigEndian.PutUint16(buffer[14:16], t.Window)
	binary.BigEndian.PutUint16(buffer[16:18], t.Checksum)
	binary.BigEndian.PutUint16(buffer[18:20], t.UrgentPointer)

	return buffer[:t.Len]
}