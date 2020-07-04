package libol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/xtaci/kcp-go/v5"
	"net"
	"time"
)

const (
	MaxBuf = 4096
	HlSize = 0x04
)

var MAGIC = []byte{0xff, 0xff}

func isControl(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	if bytes.Equal(data[:6], ZEROED[:6]) {
		return true
	}
	return false
}

type Ip4Proto struct {
	// public
	Eth  *Ether
	Vlan *Vlan
	Arp  *Arp
	Ip4  *Ipv4
	Udp  *Udp
	Tcp  *Tcp
	// private
	err   error
	frame []byte
}

func (i *Ip4Proto) Decode() error {
	data := i.frame
	if i.Eth, i.err = NewEtherFromFrame(data); i.err != nil {
		return i.err
	}
	data = data[i.Eth.Len:]
	if i.Eth.IsVlan() {
		if i.Vlan, i.err = NewVlanFromFrame(data); i.err != nil {
			return i.err
		}
		data = data[i.Vlan.Len:]
	}
	if i.Eth.IsIP4() {
		if i.Ip4, i.err = NewIpv4FromFrame(data); i.err != nil {
			return i.err
		}
		data = data[i.Ip4.Len:]
		switch i.Ip4.Protocol {
		case IpTcp:
			if i.Tcp, i.err = NewTcpFromFrame(data); i.err != nil {
				return i.err
			}
		case IpUdp:
			if i.Udp, i.err = NewUdpFromFrame(data); i.err != nil {
				return i.err
			}
		}
	} else if i.Eth.IsArp() {
		if i.Arp, i.err = NewArpFromFrame(data); i.err != nil {
			return i.err
		}
	}
	return nil
}

type FrameMessage struct {
	control bool
	action  string
	params  string
	buffer  []byte
	size    int
	total   int
	frame   []byte
	proto   *Ip4Proto
}

func NewFrameMessage() *FrameMessage {
	m := FrameMessage{
		control: false,
		action:  "",
		params:  "",
		size:    0,
		buffer:  make([]byte, HlSize+MaxBuf),
	}
	m.frame = m.buffer[HlSize:]
	m.total = len(m.frame)
	return &m
}

func (m *FrameMessage) Decode() bool {
	m.control = isControl(m.frame)
	if m.control {
		m.action = string(m.frame[6:11])
		m.params = string(m.frame[12:])
	}
	return m.control
}

func (m *FrameMessage) IsEthernet() bool {
	return !m.control
}

func (m *FrameMessage) IsControl() bool {
	return m.control
}

func (m *FrameMessage) Frame() []byte {
	return m.frame
}

func (m *FrameMessage) String() string {
	return fmt.Sprintf("control: %t, frame: %x", m.control, m.frame[:20])
}

func (m *FrameMessage) CmdAndParams() (string, string) {
	return m.action, m.params
}

func (m *FrameMessage) Append(data []byte) {
	add := len(data)
	if m.total-m.size >= add {
		copy(m.frame[m.size:], data)
		m.size += add
	}
}

func (m *FrameMessage) Size() int {
	return m.size
}

func (m *FrameMessage) SetSize(v int) {
	m.size = v
}

func (m *FrameMessage) Proto() (*Ip4Proto, error) {
	if m.proto != nil {
		return m.proto, m.proto.err
	}
	m.proto = &Ip4Proto{frame: m.frame}
	err := m.proto.Decode()
	return m.proto, err
}

type ControlMessage struct {
	control  bool
	operator string
	action   string
	params   string
}

//operator: request is '= ', and response is  ': '
//action: login, network etc.
//body: json string.
func NewControlMessage(action string, opr string, body string) *ControlMessage {
	c := ControlMessage{
		control:  true,
		action:   action,
		params:   body,
		operator: opr,
	}
	return &c
}

func (c *ControlMessage) Encode() *FrameMessage {
	p := fmt.Sprintf("%s%s%s", c.action[:4], c.operator[:2], c.params)
	frame := NewFrameMessage()
	frame.Append(ZEROED[:6])
	frame.Append([]byte(p))
	return frame
}

type Messager interface {
	Send(conn net.Conn, frame *FrameMessage) (int, error)
	Receive(conn net.Conn, max, min int) (*FrameMessage, error)
}

type StreamMessage struct {
	timeout time.Duration // ns for read and write deadline.
	block   kcp.BlockCrypt
}

func (s *StreamMessage) write(conn net.Conn, tmp []byte) (int, error) {
	if s.timeout != 0 {
		err := conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	n, err := conn.Write(tmp)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *StreamMessage) writeFull(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	size := len(buf)
	left := size - offset
	if HasLog(LOG) {
		Log("writeFull: %s %d", conn.RemoteAddr(), size)
		Log("writeFull: %s Data %x", conn.RemoteAddr(), buf)
	}
	for left > 0 {
		tmp := buf[offset:]
		if HasLog(LOG) {
			Log("writeFull: tmp %s %d", conn.RemoteAddr(), len(tmp))
		}
		n, err := s.write(conn, tmp)
		if err != nil {
			return err
		}
		if HasLog(LOG) {
			Log("writeFull: %s snd %d, size %d", conn.RemoteAddr(), n, size)
		}
		offset += n
		left = size - offset
	}
	return nil
}

func (s *StreamMessage) Send(conn net.Conn, frame *FrameMessage) (int, error) {
	frame.buffer[0] = MAGIC[0]
	frame.buffer[1] = MAGIC[1]
	binary.BigEndian.PutUint16(frame.buffer[2:4], uint16(frame.size))
	if s.block != nil {
		s.block.Encrypt(frame.frame, frame.frame)
	}
	if err := s.writeFull(conn, frame.buffer[:frame.size+4]); err != nil {
		return 0, err
	}
	return frame.size, nil
}

func (s *StreamMessage) read(conn net.Conn, tmp []byte) (int, error) {
	if s.timeout != 0 {
		err := conn.SetReadDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	n, err := conn.Read(tmp)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *StreamMessage) readFull(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	left := len(buf)
	if HasLog(LOG) {
		Log("readFull: %s %d", conn.RemoteAddr(), len(buf))
	}
	for left > 0 {
		tmp := make([]byte, left)
		n, err := s.read(conn, tmp)
		if err != nil {
			return err
		}
		copy(buf[offset:], tmp)
		offset += n
		left -= n
	}
	if HasLog(LOG) {
		Log("readFull: Data %s %x", conn.RemoteAddr(), buf)
	}
	return nil
}

func (s *StreamMessage) Receive(conn net.Conn, max, min int) (*FrameMessage, error) {
	frame := NewFrameMessage()
	h := frame.buffer[:4]
	if err := s.readFull(conn, h); err != nil {
		return nil, err
	}
	if !bytes.Equal(h[:2], MAGIC[:2]) {
		return nil, NewErr("%s: wrong magic", conn.RemoteAddr())
	}
	size := int(binary.BigEndian.Uint16(h[2:4]))
	if size > max || size < min {
		return nil, NewErr("%s: wrong size %d", conn.RemoteAddr(), size)
	}
	tmp := frame.buffer[4 : 4+size]
	if err := s.readFull(conn, tmp); err != nil {
		return nil, err
	}
	if s.block != nil {
		s.block.Decrypt(tmp, tmp)
	}
	frame.size = size
	frame.frame = tmp
	return frame, nil
}

type DataGramMessage struct {
	timeout time.Duration // ns for read and write deadline
	block   kcp.BlockCrypt
}

func (s *DataGramMessage) Send(conn net.Conn, frame *FrameMessage) (int, error) {
	frame.buffer[0] = MAGIC[0]
	frame.buffer[1] = MAGIC[1]
	binary.BigEndian.PutUint16(frame.buffer[2:4], uint16(frame.size))
	if s.block != nil {
		s.block.Encrypt(frame.frame, frame.frame)
	}
	if HasLog(DEBUG) {
		Debug("DataGramMessage.Send: %s %x", conn.RemoteAddr(), frame)
	}
	if s.timeout != 0 {
		err := conn.SetWriteDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return 0, err
		}
	}
	if _, err := conn.Write(frame.buffer[:4+frame.size]); err != nil {
		return 0, err
	}
	return frame.size, nil
}

func (s *DataGramMessage) Receive(conn net.Conn, max, min int) (*FrameMessage, error) {
	frame := NewFrameMessage()
	if HasLog(DEBUG) {
		Debug("DataGramMessage.Receive %s %d", conn.RemoteAddr(), s.timeout)
	}
	if s.timeout != 0 {
		err := conn.SetReadDeadline(time.Now().Add(s.timeout))
		if err != nil {
			return nil, err
		}
	}
	n, err := conn.Read(frame.buffer)
	if err != nil {
		return nil, err
	}
	if HasLog(DEBUG) {
		Debug("DataGramMessage.Receive: %s %x", conn.RemoteAddr(), frame.buffer)
	}
	if n <= 4 {
		return nil, NewErr("%s: small frame", conn.RemoteAddr())
	}
	if !bytes.Equal(frame.buffer[:2], MAGIC[:2]) {
		return nil, NewErr("%s: wrong magic", conn.RemoteAddr())
	}
	size := int(binary.BigEndian.Uint16(frame.buffer[2:4]))
	if size > max || size < min {
		return nil, NewErr("%s: wrong size %d", conn.RemoteAddr(), size)
	}
	tmp := frame.buffer[4 : 4+size]
	if s.block != nil {
		s.block.Decrypt(tmp, tmp)
	}
	frame.size = size
	frame.frame = tmp
	return frame, nil
}
