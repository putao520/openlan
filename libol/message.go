package libol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	MAXBUF = 4096
	HSIZE  = 0x04
)

func GetHeaderLen() int {
	return HSIZE
}

var MAGIC = []byte{0xff, 0xff}

func GetMagic() []byte {
	return MAGIC
}

func IsControl(data []byte) bool {
	if bytes.Equal(data[:6], ZEROED[:6]) {
		return true
	}
	return false
}

type FrameMessage struct {
	control bool
	action  string
	params  string
	frame   []byte
	raw     []byte
}

func NewFrameMessage(data []byte) *FrameMessage {
	m := FrameMessage{
		control: false,
		action:  "",
		params:  "",
		raw:     data,
	}
	m.Decode()
	return &m
}

func (m *FrameMessage) Decode() bool {
	m.control = IsControl(m.raw)
	if m.control {
		m.action = string(m.raw[6:11])
		m.params = string(m.raw[12:])
	} else {
		m.frame = m.raw
	}
	return m.control
}

func (m *FrameMessage) IsControl() bool {
	return m.control
}

func (m *FrameMessage) Data() []byte {
	return m.frame
}

func (m *FrameMessage) String() string {
	return fmt.Sprintf("control: %t, raw: %x", m.control, m.raw[:20])
}

func (m *FrameMessage) CmdAndParams() (string, string) {
	return m.action, m.params
}

type ControlMessage struct {
	control bool
	opr     string
	action  string
	params  string
}

//opr: request is '= ', and response is  ': '
//action: login, network etc.
//body: json string.
func NewControlMessage(action string, opr string, body string) *ControlMessage {
	c := ControlMessage{
		control: true,
		action:  action,
		params:  body,
		opr:     opr,
	}

	return &c
}

func (c *ControlMessage) Encode() []byte {
	p := fmt.Sprintf("%s%s%s", c.action[:4], c.opr[:2], c.params)
	return append(ZEROED[:6], p...)
}

func readFull(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	left := len(buf)
	Log("readFull: %s %d", conn.RemoteAddr(), len(buf))
	for left > 0 {
		tmp := make([]byte, left)
		n, err := conn.Read(tmp)
		if err != nil {
			return err
		}
		copy(buf[offset:], tmp)
		offset += n
		left -= n
	}
	Log("readFull: Data %s %x", conn.RemoteAddr(), buf)
	return nil
}

func writeFull(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	size := len(buf)
	left := size - offset
	Log("writeFull: %s %d", conn.RemoteAddr(), size)
	Log("writeFull: %s Data %x", conn.RemoteAddr(), buf)
	for left > 0 {
		tmp := buf[offset:]
		Log("writeFull: tmp %s %d", conn.RemoteAddr(), len(tmp))
		n, err := conn.Write(tmp)
		if err != nil {
			return err
		}
		Log("writeFull: %s snd %d, size %d", conn.RemoteAddr(), n, size)
		offset += n
		left = size - offset
	}
	return nil
}

type Messager interface {
	Send(conn net.Conn, data []byte) (int, error)
	Receive(conn net.Conn, data []byte, max, min int) (int, error)
}

type StreamMessage struct {
	//TODO
}

func (s *StreamMessage) Send(conn net.Conn, data []byte) (int, error) {
	size := len(data)
	buf := make([]byte, HSIZE+size)
	copy(buf[0:2], MAGIC)
	binary.BigEndian.PutUint16(buf[2:4], uint16(size))
	copy(buf[HSIZE:], data)

	if err := writeFull(conn, buf); err != nil {
		return 0, err
	}
	return size, nil
}

func (s *StreamMessage) Receive(conn net.Conn, data []byte, max, min int) (int, error) {
	hl := GetHeaderLen()
	buf := make([]byte, hl+max)
	h := buf[:hl]
	if err := readFull(conn, h); err != nil {
		return -1, err
	}
	magic := GetMagic()
	if !bytes.Equal(h[0:2], magic) {
		return -1, NewErr("%s: wrong magic", conn.RemoteAddr())
	}
	size := binary.BigEndian.Uint16(h[2:4])
	if int(size) > max || int(size) < min {
		return -1, NewErr("%s: wrong size(%d)", conn.RemoteAddr(), size)
	}
	d := buf[hl : hl+int(size)]
	if err := readFull(conn, d); err != nil {
		return -1, err
	}
	copy(data, d)
	return len(d), nil
}

type DataGramMessage struct {
	//TODO
}

func (s *DataGramMessage) Send(conn net.Conn, data []byte) (int, error) {
	size := len(data)
	buf := make([]byte, HSIZE+size)
	copy(buf[0:2], MAGIC)
	binary.BigEndian.PutUint16(buf[2:4], uint16(size))
	copy(buf[HSIZE:], data)

	Log("DataGramMessage.Send: %s %x", conn.RemoteAddr(), data)
	if _, err := conn.Write(buf); err != nil {
		return 0, err
	}
	return size, nil
}

func (s *DataGramMessage) Receive(conn net.Conn, data []byte, max, min int) (int, error) {
	hl := GetHeaderLen()
	buf := make([]byte, hl+max)

	n, err := conn.Read(buf)
	if err != nil {
		return -1, err
	}
	Log("DataGramMessage.Receive: %s %x", conn.RemoteAddr(), buf[:n])
	if n <= hl {
		return -1, NewErr("%s: small frame", conn.RemoteAddr())
	}
	magic := GetMagic()
	if !bytes.Equal(buf[0:2], magic) {
		return -1, NewErr("%s: wrong magic", conn.RemoteAddr())
	}
	size := binary.BigEndian.Uint16(buf[2:4])
	if int(size) > max || int(size) < min {
		return -1, NewErr("%s: wrong size(%d)", conn.RemoteAddr(), size)
	}
	d := buf[hl : hl+int(size)]
	copy(data, d)
	return len(d), nil
}
