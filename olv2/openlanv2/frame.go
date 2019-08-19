package openlanv2

import (
	"fmt"
	"encoding/binary"
)

var (
	ZEROMAC = []byte{0x00,0x00,0x00,0x00,0x00,0x00}
	BROADCASTMAC = []byte{0xff,0xff,0xff,0xff,0xff,0xff}
	ETHLEN = 16 //6+6+2+2
)

var (
	ETH_ARP  = 0x0806
	ETH_VLAN = 0x8100
	ETH_IPV4 = 0x0800
)

func EthType(data []byte) uint16 {
	return binary.BigEndian.Uint16(data[12:14])
}

func DstAddr(data []byte) []byte {
	return data[0:6]
}

func SrcAddr(data []byte) []byte {
	return data[6:12]
}

func EthAddrStr(data []byte) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", 
			data[0], data[1], data[2], data[3], data[4], data[5])
}