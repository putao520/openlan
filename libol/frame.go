package libol

import (
	"encoding/binary"
)

type Frame struct {
	Data []byte
	Eth  *Ether
}

func NewFrame(data []byte) (f *Frame) {
	f = &Frame{
		Data: data,
	}

	eth, err := NewEtherFromFrame(f.Data)
	if err == nil {
		f.Eth = eth
	} else {
		Error("NewFrame: %s", err)
	}

	return
}

func (f *Frame) EthType() uint16 {
	return binary.BigEndian.Uint16(f.Data[12:14])
}

func (f *Frame) EthData() []byte {
	return f.Data[14:]
}

func (f *Frame) DstAddr() []byte {
	return f.Data[0:6]
}

func (f *Frame) SrcAddr() []byte {
	return f.Data[6:12]
}

func (f *Frame) EthParse() (uint16, []byte) {
	return f.EthType(), f.EthData()
}
