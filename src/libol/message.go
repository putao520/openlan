package libol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/xtaci/kcp-go/v5"
	"net"
	"sync"
	"time"
)

const (
	MaxBuf = 4096
	HlSize = 0x04
)

var MAGIC = []byte{0xff, 0xff}

const (
	LoginReq     = "logi= "
	LoginResp    = "logi: "
	NeighborReq  = "neig= "
	NeighborResp = "neig: "
	IpAddrReq    = "ipad= "
	IpAddrResp   = "ipad: "
	LeftReq      = "left= "
	SignReq      = "sign= "
	PingReq      = "ping= "
	PongResp     = "pong: "
)

func isControl(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	if bytes.Equal(data[:6], EthZero[:6]) {
		return true
	}
	return false
}

type FrameProto struct {
	// public
	Eth   *Ether
	Vlan  *Vlan
	Arp   *Arp
	Ip4   *Ipv4
	Udp   *Udp
	Tcp   *Tcp
	Err   error
	Frame []byte
}

func (i *FrameProto) Decode() error {
	data := i.Frame
	if i.Eth, i.Err = NewEtherFromFrame(data); i.Err != nil {
		return i.Err
	}
	data = data[i.Eth.Len:]
	if i.Eth.IsVlan() {
		if i.Vlan, i.Err = NewVlanFromFrame(data); i.Err != nil {
			return i.Err
		}
		data = data[i.Vlan.Len:]
	}
	switch i.Eth.Type {
	case EthIp4:
		if i.Ip4, i.Err = NewIpv4FromFrame(data); i.Err != nil {
			return i.Err
		}
		data = data[i.Ip4.Len:]
		switch i.Ip4.Protocol {
		case IpTcp:
			if i.Tcp, i.Err = NewTcpFromFrame(data); i.Err != nil {
				return i.Err
			}
		case IpUdp:
			if i.Udp, i.Err = NewUdpFromFrame(data); i.Err != nil {
				return i.Err
			}
		}
	case EthArp:
		if i.Arp, i.Err = NewArpFromFrame(data); i.Err != nil {
			return i.Err
		}
	}
	return nil
}

type FrameMessage struct {
	seq     uint64
	control bool
	action  string
	params  []byte
	buffer  []byte
	size    int
	total   int
	frame   []byte
	proto   *FrameProto
	header  *FrameLLHeader
}

func NewFrameMessage() *FrameMessage {
	m := FrameMessage{
		params: make([]byte, 0, 2),
		buffer: make([]byte, HlSize+MaxBuf),
	}
	m.frame = m.buffer[HlSize:]
	m.total = len(m.frame)
	return &m
}

func NewFrameMessageFromBytes(buffer []byte) *FrameMessage {
	m := FrameMessage{
		params: make([]byte, 0, 2),
		buffer: buffer,
	}
	m.frame = m.buffer[HlSize:]
	m.total = len(m.frame)
	m.size = len(m.frame)
	return &m
}

func (m *FrameMessage) Decode() bool {
	m.control = isControl(m.frame)
	if m.control {
		m.action = string(m.frame[6:12])
		m.params = m.frame[12:]
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

func (m *FrameMessage) Action() string {
	return m.action
}

func (m *FrameMessage) CmdAndParams() (string, []byte) {
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

func (m *FrameMessage) Proto() (*FrameProto, error) {
	if m.proto != nil {
		return m.proto, m.proto.Err
	}
	m.proto = &FrameProto{Frame: m.frame}
	err := m.proto.Decode()
	return m.proto, err
}

const (
	FmsNew   = 0
	FmsWrFin = 1
	FmsRdFin = 2
)

type FrameLLNode struct {
	next *FrameLLNode
	data *FrameMessage
}

type FrameLLHeader struct {
	lock  *sync.RWMutex
	state int
	tail  *FrameLLNode
	next  *FrameLLNode
}

func (l *FrameLLHeader) Have(state int) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.state == state
}

func (l *FrameLLHeader) State(state int) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.state = state
}

func (l *FrameLLHeader) Next() *FrameLLNode {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.next
}

func (l *FrameLLHeader) Add(data *FrameMessage) {
	l.lock.Lock()
	defer l.lock.Unlock()
	next := &FrameLLNode{data: data}
	if l.tail == nil {
		l.tail = next
	} else if l.tail != nil {
		l.tail.next = next
		l.tail = next
	}
}

type ControlMessage struct {
	seq      uint64
	control  bool
	operator string
	action   string
	params   []byte
}

func NewControlFrame(action string, body []byte) *FrameMessage {
	m := NewControlMessage(action[:4], action[4:], body)
	return m.Encode()
}

//operator: request is '= ', and response is  ': '
//action: login, network etc.
//body: json string.
func NewControlMessage(action, opr string, body []byte) *ControlMessage {
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
	frame.Append(EthZero[:6])
	frame.Append([]byte(p))
	return frame
}

type Messager interface {
	Send(conn net.Conn, frame *FrameMessage) (int, error)
	Receive(conn net.Conn, max, min int) (*FrameMessage, error)
	Flush()
}

type StreamMessagerImpl struct {
	timeout time.Duration // ns for read and write deadline.
	block   kcp.BlockCrypt
	buffer  []byte
	bufSize int // default is (1518 + 20+20+14) * 8
}

func (s *StreamMessagerImpl) Flush() {
	s.buffer = nil
}

func (s *StreamMessagerImpl) write(conn net.Conn, tmp []byte) (int, error) {
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

func (s *StreamMessagerImpl) writeX(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	size := len(buf)
	left := size - offset
	if HasLog(LOG) {
		Log("StreamMessagerImpl.writeX: %s %d", conn.RemoteAddr(), size)
		Log("StreamMessagerImpl.writeX: %s Data %x", conn.RemoteAddr(), buf)
	}
	for left > 0 {
		tmp := buf[offset:]
		if HasLog(LOG) {
			Log("StreamMessagerImpl.writeX: tmp %s %d", conn.RemoteAddr(), len(tmp))
		}
		n, err := s.write(conn, tmp)
		if err != nil {
			return err
		}
		if HasLog(LOG) {
			Log("StreamMessagerImpl.writeX: %s snd %d, size %d", conn.RemoteAddr(), n, size)
		}
		offset += n
		left = size - offset
	}
	return nil
}

func (s *StreamMessagerImpl) Send(conn net.Conn, frame *FrameMessage) (int, error) {
	frame.buffer[0] = MAGIC[0]
	frame.buffer[1] = MAGIC[1]
	binary.BigEndian.PutUint16(frame.buffer[2:4], uint16(frame.size))
	if s.block != nil {
		s.block.Encrypt(frame.frame, frame.frame)
	}
	if err := s.writeX(conn, frame.buffer[:frame.size+4]); err != nil {
		return 0, err
	}
	return frame.size, nil
}

func (s *StreamMessagerImpl) read(conn net.Conn, tmp []byte) (int, error) {
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

//340Mib
func (s *StreamMessagerImpl) readX(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	left := len(buf)
	if HasLog(LOG) {
		Log("StreamMessagerImpl.readX: %s %d", conn.RemoteAddr(), len(buf))
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
		Log("StreamMessagerImpl.readX: Data %s %x", conn.RemoteAddr(), buf)
	}
	return nil
}

func (s *StreamMessagerImpl) fetchFrame(tmp []byte, min int) (*FrameMessage, error) {
	ts := len(tmp)
	if ts < min {
		return nil, nil
	}
	if !bytes.Equal(tmp[:2], MAGIC[:2]) {
		return nil, NewErr("wrong magic")
	}
	ps := binary.BigEndian.Uint16(tmp[2:4])
	fs := int(ps) + HlSize
	if ts >= fs {
		s.buffer = tmp[fs:]
		if s.block != nil {
			s.block.Decrypt(tmp[HlSize:fs], tmp[HlSize:fs])
		}
		return NewFrameMessageFromBytes(tmp[:fs]), nil
	}
	return nil, nil
}

// 430Mib
func (s *StreamMessagerImpl) Receive(conn net.Conn, max, min int) (*FrameMessage, error) {
	frame, err := s.fetchFrame(s.buffer, min)
	if err != nil {
		return nil, err
	}
	if frame != nil { // firstly, check buffer has messages.
		return frame, nil
	}
	if s.bufSize == 0 {
		s.bufSize = 12576 // 1572 * 8
	}
	bs := len(s.buffer)
	tmp := make([]byte, s.bufSize)
	if bs > 0 {
		copy(tmp[:bs], s.buffer[:bs])
	}
	for { // loop forever until socket error or find one message.
		rn, err := conn.Read(tmp[bs:])
		if err != nil {
			return nil, err
		}
		rs := bs + rn
		frame, err := s.fetchFrame(tmp[:rs], min)
		if err != nil {
			return nil, err
		}
		if frame != nil {
			return frame, nil
		}
		// If notFound message, continue to read.
		bs = rs
	}
}

type PacketMessagerImpl struct {
	timeout time.Duration // ns for read and write deadline
	block   kcp.BlockCrypt
}

func (s *PacketMessagerImpl) Flush() {
	//TODO
}

func (s *PacketMessagerImpl) Send(conn net.Conn, frame *FrameMessage) (int, error) {
	frame.buffer[0] = MAGIC[0]
	frame.buffer[1] = MAGIC[1]
	binary.BigEndian.PutUint16(frame.buffer[2:4], uint16(frame.size))
	if s.block != nil {
		s.block.Encrypt(frame.frame, frame.frame)
	}
	if HasLog(DEBUG) {
		Debug("PacketMessagerImpl.Send: %s %x", conn.RemoteAddr(), frame)
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

func (s *PacketMessagerImpl) Receive(conn net.Conn, max, min int) (*FrameMessage, error) {
	frame := NewFrameMessage()
	if HasLog(DEBUG) {
		Debug("PacketMessagerImpl.Receive %s %d", conn.RemoteAddr(), s.timeout)
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
		Debug("PacketMessagerImpl.Receive: %s %x", conn.RemoteAddr(), frame.buffer)
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
