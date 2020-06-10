package libol

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync"
	"time"
)

const (
	CL_INIT       = 0x00
	CL_CONNECTED  = 0x01
	CL_UNAUTH     = 0x02
	CL_AUEHED     = 0x03
	CL_CONNECTING = 0x04
	CL_TERMINAL   = 0x05
	CL_CLOSED     = 0x06
)

type ClientSts struct {
	TxOkay  uint64
	RxOkay  uint64
	TxError uint64
	Dropped uint64
}

type ClientListener struct {
	OnClose     func(client SocketClient) error
	OnConnected func(client SocketClient) error
	OnStatus    func(client SocketClient, old, new uint8)
}

type SocketClient interface {
	LocalAddr() string
	Connect() error
	Close()
	WriteMsg(data []byte) error
	ReadMsg(data []byte) (int, error)
	WriteReq(action string, body string) error
	WriteResp(action string, body string) error
	State() string
	UpTime() int64
	String() string
	Terminal()
	Private() interface{}
	SetPrivate(v interface{})
	Status() uint8
	SetStatus(v uint8)
	MaxSize() int
	SetMaxSize(value int)
	MinSize() int
	IsOk() bool
	Have(status uint8) bool
	Addr() string
	SetAddr(addr string)
	Sts() ClientSts
	SetListener(listener ClientListener)
}

func readFull(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	left := len(buf)
	Log("readFull: %d", len(buf))
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
	Log("readFull: Data %x", buf)
	return nil
}

func writeFull(conn net.Conn, buf []byte) error {
	if conn == nil {
		return NewErr("connection is nil")
	}
	offset := 0
	size := len(buf)
	left := size - offset
	Log("writeFull: %d", size)
	Log("writeFull: Data %x", buf)
	for left > 0 {
		tmp := buf[offset:]
		Log("writeFull: tmp %d", len(tmp))
		n, err := conn.Write(tmp)
		if err != nil {
			return err
		}
		Log("writeFull: snd %d, size %d", n, size)
		offset += n
		left = size - offset
	}
	return nil
}

type connWrapper struct {
	conn    net.Conn
	sts     ClientSts
	maxSize int
	minSize int
	connect func() error
}

func (t *connWrapper) String() string {
	if t.conn != nil {
		return t.conn.RemoteAddr().String()
	}
	return "unknown"
}

func (t *connWrapper) IsOk() bool {
	return t.conn != nil
}

func (t *connWrapper) WriteMsg(data []byte) error {
	if err := t.connect(); err != nil {
		t.sts.Dropped++
		return err
	}
	buf := BuildMessage(data)
	if err := writeFull(t.conn, buf); err != nil {
		t.sts.TxError++
		return err
	}
	t.sts.TxOkay += uint64(len(data))
	return nil
}

func (t *connWrapper) ReadMsg(data []byte) (int, error) {
	Log("connWrapper.ReadMsg: %s", t)

	if !t.IsOk() {
		return -1, NewErr("%s: not okay", t)
	}

	hl := GetHeaderLen()
	buffer := make([]byte, hl+t.maxSize)
	h := buffer[:hl]
	if err := readFull(t.conn, h); err != nil {
		return -1, err
	}
	magic := GetMagic()
	if !bytes.Equal(h[0:2], magic) {
		return -1, NewErr("%s: wrong magic", t)
	}

	size := binary.BigEndian.Uint16(h[2:4])
	if int(size) > t.maxSize || int(size) < t.minSize {
		return -1, NewErr("%s: wrong size(%d)", t, size)
	}
	d := buffer[hl : hl+int(size)]
	if err := readFull(t.conn, d); err != nil {
		return -1, err
	}

	copy(data, d)
	t.sts.RxOkay += uint64(size)

	return len(d), nil
}

func (t *connWrapper) WriteReq(action string, body string) error {
	m := NewControlMessage(action, "= ", body)
	data := m.Encode()
	Log("connWrapper.WriteReq: %d %s", len(data), data[6:])
	return t.WriteMsg(data)
}

func (t *connWrapper) WriteResp(action string, body string) error {
	m := NewControlMessage(action, ": ", body)
	data := m.Encode()
	Log("connWrapper.WriteResp: %d %s", len(data), data[6:])
	return t.WriteMsg(data)
}

type socketClient struct {
	connWrapper
	lock     sync.RWMutex
	listener ClientListener
	addr     string
	NewTime  int64
	private  interface{}
	//sts      ClientSts
	//maxSize  int
	//minSize  int
	status uint8
}

func (s *socketClient) State() string {
	switch s.Status() {
	case CL_INIT:
		return "initialized"
	case CL_CONNECTED:
		return "connected"
	case CL_UNAUTH:
		return "unauthenticated"
	case CL_AUEHED:
		return "authenticated"
	case CL_CLOSED:
		return "closed"
	case CL_CONNECTING:
		return "connecting"
	case CL_TERMINAL:
		return "terminal"
	}
	return ""
}

func (s *socketClient) Status() uint8 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.status
}

func (s *socketClient) UpTime() int64 {
	return time.Now().Unix() - s.NewTime
}

func (s *socketClient) Addr() string {
	return s.addr
}

func (s *socketClient) SetAddr(addr string) {
	s.addr = addr
}

func (s *socketClient) String() string {
	return s.Addr()
}

func (s *socketClient) Private() interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.private
}

func (s *socketClient) SetPrivate(v interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.private = v
}

func (s *socketClient) MaxSize() int {
	return s.maxSize
}

func (s *socketClient) SetMaxSize(value int) {
	s.maxSize = value
}

func (s *socketClient) MinSize() int {
	return s.minSize
}

func (s *socketClient) Have(state uint8) bool {
	return s.Status() == state
}

func (s *socketClient) Sts() ClientSts {
	return s.sts
}

func (s *socketClient) SetListener(listener ClientListener) {
	s.listener = listener
}

// Socket Server

type ServerSts struct {
	RxCount  int64
	TxCount  int64
	DrpCount int64
	AcpCount int64
	ClsCount int64
}

type ServerListener struct {
	OnClient func(client SocketClient) error
	OnClose  func(client SocketClient) error
	ReadAt   func(client SocketClient, p []byte) error
}

type ReadClient func(client SocketClient, p []byte) error

type SocketServer interface {
	Listen() (err error)
	Close()
	Accept()
	OffClient(client SocketClient)
	Loop(call ServerListener)
	Read(client SocketClient, ReadAt ReadClient)
	String() string
	Addr() string
	Sts() ServerSts
}

type socketServer struct {
	lock       sync.RWMutex
	sts        ServerSts
	addr       string
	maxClient  int
	clients    map[SocketClient]bool
	onClients  chan SocketClient
	offClients chan SocketClient
	close      func()
}

func (t *socketServer) OffClient(client SocketClient) {
	Warn("socketServer.OffClient %s", client)
	if client != nil {
		t.offClients <- client
	}
}

func (t *socketServer) doOnClient(call ServerListener, client SocketClient) {
	Debug("socketServer.doOnClient: %s", client.Addr())
	t.clients[client] = true
	if call.OnClient != nil {
		_ = call.OnClient(client)
		if call.ReadAt != nil {
			go t.Read(client, call.ReadAt)
		}
	}
}

func (t *socketServer) doOffClient(call ServerListener, client SocketClient) {
	Debug("socketServer.doOffClient: %s", client.Addr())
	if ok := t.clients[client]; ok {
		t.sts.ClsCount++
		if call.OnClose != nil {
			_ = call.OnClose(client)
		}
		client.Close()
		delete(t.clients, client)
	}
}

func (t *socketServer) Loop(call ServerListener) {
	Debug("socketServer.Loop")
	defer t.close()
	for {
		select {
		case client := <-t.onClients:
			t.doOnClient(call, client)
		case client := <-t.offClients:
			t.doOffClient(call, client)
		}
	}
}

func (t *socketServer) Read(client SocketClient, ReadAt ReadClient) {
	Log("socketServer.Read: %s", client.Addr())
	data := make([]byte, MAXBUF)
	for {
		length, err := client.ReadMsg(data)
		if err != nil {
			Error("socketServer.Read: %s", err)
			t.OffClient(client)
			break
		}
		if length <= 0 {
			continue
		}
		t.sts.RxCount++
		Log("socketServer.Read: length: %d ", length)
		Log("socketServer.Read: data  : %x", data[:length])
		if err := ReadAt(client, data[:length]); err != nil {
			Error("socketServer.Read: readAt %s", err)
			break
		}
	}
}

func (t *socketServer) Close() {
	if t.close != nil {
		t.close()
	}
}

func (t *socketServer) Addr() string {
	return t.addr
}

func (t *socketServer) String() string {
	return t.Addr()
}

func (t *socketServer) Sts() ServerSts {
	return t.sts
}
