package openlanv1

import (
    "fmt"
    "encoding/binary"
)

var (
    ZEROMAC = []byte{0x00,0x00,0x00,0x00,0x00,0x00}
    BROADCASTMAC = []byte{0xff,0xff,0xff,0xff,0xff,0xff}
)

var (
    ETH_ARP  = 0x0806
    ETH_VLAN = 0x8100
    ETH_IPV4 = 0x0800
)

type Frame struct {
    Data []byte
}

func NewFrame(data []byte) (this *Frame) {
    this = &Frame{
        Data: make([]byte, len(data)),
    }

    copy(this.Data, data)
    return 
}

func (this *Frame) EthType() uint16 {
    return binary.BigEndian.Uint16(this.Data[12:14])
}

func (this *Frame) EthData() {
    return this.Data[14:]
}

func (this *Frame) DstAddr() []byte {
    return this.Data[0:6]
}

func (this *Frame) SrcAddr() []byte {
    return this.Data[6:12]
}