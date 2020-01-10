package libol

import (
	"encoding/binary"
	"fmt"
)

const (
	IPV4_VER = 0x04
	IPV6_VER = 0x06
)

const (
	IPPROTO_ICMP = 0x01
	IPPROTO_IGMP = 0x02
	IPPROTO_IPIP = 0x04
	IPPROTO_TCP  = 0x06
	IPPROTO_UDP  = 0x11
	IPPROTO_ESP  = 0x32
	IPPROTO_AH   = 0x33
	IPPROTO_OSPF = 0x59
	IPPROTO_PIM  = 0x67
	IPPROTO_VRRP = 0x70
	IPPROTO_ISIS = 0x7c
)

func IpProto2Str(proto uint8) string {
	switch proto {
	case IPPROTO_ICMP:
		return "icmp"
	case IPPROTO_IGMP:
		return "igmp"
	case IPPROTO_IPIP:
		return "ipip"
	case IPPROTO_ESP:
		return "esp"
	case IPPROTO_AH:
		return "ah"
	case IPPROTO_OSPF:
		return "ospf"
	case IPPROTO_ISIS:
		return "isis"
	case IPPROTO_UDP:
		return "udp"
	case IPPROTO_TCP:
		return "tcp"
	case IPPROTO_PIM:
		return "pim"
	case IPPROTO_VRRP:
		return "vrrp"
	default:
		return fmt.Sprintf("%02x", proto)
	}
}

const IPV4_LEN = 20

type Ipv4 struct {
	Version        uint8 //4bite v4: 0100, v6: 0110
	HeaderLen      uint8 //4bit 15*4
	ToS            uint8 //Type of Service
	TotalLen       uint16
	Identifier     uint16
	Flag           uint16 //3bit Z|DF|MF
	Offset         uint16 //13bit Fragment offset
	ToL            uint8  //Time of Live
	Protocol       uint8
	HeaderChecksum uint16 //Header Checksum
	Source         []byte
	Destination    []byte
	Options        uint32 //Reserved
	Len            int
}

func NewIpv4() (i *Ipv4) {
	i = &Ipv4{
		Version:        0x04,
		HeaderLen:      0x05,
		ToS:            0,
		TotalLen:       0,
		Identifier:     0,
		Flag:           0,
		Offset:         0,
		ToL:            0xff,
		Protocol:       0,
		HeaderChecksum: 0,
		Options:        0,
		Len:            IPV4_LEN,
	}
	return
}

func NewIpv4FromFrame(frame []byte) (i *Ipv4, err error) {
	i = NewIpv4()
	err = i.Decode(frame)
	return
}

func (i *Ipv4) Decode(frame []byte) error {
	if len(frame) < IPV4_LEN {
		return NewErr("Ipv4.Decode: too small header: %d", len(frame))
	}

	h := uint8(frame[0])
	i.Version = h >> 4
	i.HeaderLen = h & 0x0f
	i.ToS = uint8(frame[1])
	i.TotalLen = binary.BigEndian.Uint16(frame[2:4])
	i.Identifier = binary.BigEndian.Uint16(frame[4:6])
	f := binary.BigEndian.Uint16(frame[6:8])
	i.Offset = f & 0x1fFf
	i.Flag = f >> 13
	i.ToL = uint8(frame[8])
	i.Protocol = uint8(frame[9])
	i.HeaderChecksum = binary.BigEndian.Uint16(frame[10:12])
	if !i.IsIP4() {
		return NewErr("Ipv4.Decode: not right ipv4 version: 0x%x", i.Version)
	}
	i.Source = frame[12:16]
	i.Destination = frame[16:20]

	return nil
}

func (i *Ipv4) Encode() []byte {
	buffer := make([]byte, 32)

	h := uint8((i.Version << 4) | i.HeaderLen)
	buffer[0] = h
	buffer[1] = i.ToS
	binary.BigEndian.PutUint16(buffer[2:4], i.TotalLen)
	binary.BigEndian.PutUint16(buffer[4:6], i.Identifier)
	f := uint16((i.Flag << 13) | i.Offset)
	binary.BigEndian.PutUint16(buffer[6:8], f)
	buffer[8] = i.ToL
	buffer[9] = i.Protocol
	binary.BigEndian.PutUint16(buffer[10:12], i.HeaderChecksum)
	copy(buffer[12:16], i.Source[:4])
	copy(buffer[16:20], i.Destination[:4])

	return buffer[:i.Len]
}

func (i *Ipv4) IsIP4() bool {
	return i.Version == IPV4_VER
}