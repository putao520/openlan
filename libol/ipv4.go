package libol

import (
	"bytes"
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
	FragOffset     uint16 //13bit Fragment offset
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
		FragOffset:     0,
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
	var err error

	if len(frame) < IPV4_LEN {
		return Errer("Ipv4.Decode: too small header: %d", len(frame))
	}

	reader := bytes.NewReader(frame)

	h := uint8(0)
	err = binary.Read(reader, binary.BigEndian, &h)
	i.Version = h >> 4
	i.HeaderLen = h & 0x0f

	err = binary.Read(reader, binary.BigEndian, &i.ToS)
	err = binary.Read(reader, binary.BigEndian, &i.TotalLen)
	err = binary.Read(reader, binary.BigEndian, &i.Identifier)
	err = binary.Read(reader, binary.BigEndian, &i.FragOffset)
	i.Flag = i.FragOffset >> 13
	err = binary.Read(reader, binary.BigEndian, &i.ToL)
	err = binary.Read(reader, binary.BigEndian, &i.Protocol)
	err = binary.Read(reader, binary.BigEndian, &i.HeaderChecksum)

	if !i.IsIP4() {
		return Errer("Ipv4.Decode: not right ipv4 version: 0x%x", i.Version)
	}

	i.Source = make([]byte, 4)
	err = binary.Read(reader, binary.BigEndian, i.Source)
	i.Destination = make([]byte, 4)
	err = binary.Read(reader, binary.BigEndian, i.Destination)

	return err
}

func (i *Ipv4) Encode() []byte {
	writer := new(bytes.Buffer)

	h := (i.Version << 4) | i.HeaderLen
	_ = binary.Write(writer, binary.BigEndian, &h)
	_ = binary.Write(writer, binary.BigEndian, &i.ToS)
	_ = binary.Write(writer, binary.BigEndian, &i.TotalLen)
	_ = binary.Write(writer, binary.BigEndian, &i.Identifier)
	f := uint16(i.Flag <<13 | i.FragOffset)
	_ = binary.Write(writer, binary.BigEndian, &f)
	_ = binary.Write(writer, binary.BigEndian, &i.ToL)
	_ = binary.Write(writer, binary.BigEndian, &i.Protocol)
	//TODO checksum.
	_ = binary.Write(writer, binary.BigEndian, &i.HeaderChecksum)
	_ = binary.Write(writer, binary.BigEndian, i.Source[:4])
	_ = binary.Write(writer, binary.BigEndian, i.Destination[:4])

	return writer.Bytes()
}

func (i *Ipv4) IsIP4() bool {
	return i.Version == IPV4_VER
}
