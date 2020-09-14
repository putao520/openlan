package libol

import (
	"net"
	"sync"
	"time"
)

const (
	ClInit       = 0x00
	ClConnected  = 0x01
	ClUnAuth     = 0x02
	ClAuth       = 0x03
	ClConnecting = 0x04
	ClTerminal   = 0x05
	ClClosed     = 0x06
)

const (
	CsSendOkay  = "send"
	CsRecvOkay  = "recv"
	CsSendError = "error"
	CsDropped   = "dropped"
)

type ClientListener struct {
	OnClose     func(client SocketClient) error
	OnConnected func(client SocketClient) error
	OnStatus    func(client SocketClient, old, new uint8)
}

type SocketClient interface {
	LocalAddr() string
	RemoteAddr() string
	Connect() error
	Close()
	WriteMsg(frame *FrameMessage) error
	ReadMsg() (*FrameMessage, error)
	State() string
	UpTime() int64
	AliveTime() int64
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
	Address() string
	Statistics() map[string]int64
	SetListener(listener ClientListener)
	SetTimeout(v int64)
	Out() *SubLogger
}

type StreamSocket struct {
	message    Messager
	connection net.Conn
	statistics *SafeStrInt64
	maxSize    int
	minSize    int
	out        *SubLogger
	address    string
}

func (t *StreamSocket) String() string {
	if t.connection != nil {
		return t.connection.RemoteAddr().String()
	}
	return t.address
}

func (t *StreamSocket) IsOk() bool {
	return t.connection != nil
}

func (t *StreamSocket) WriteMsg(frame *FrameMessage) error {
	if !t.IsOk() {
		t.statistics.Add(CsDropped, 1)
		return NewErr("%s not okay", t)
	}
	if t.message == nil { // default is stream message
		t.message = &StreamMessagerImpl{}
	}
	size, err := t.message.Send(t.connection, frame)
	if err != nil {
		t.statistics.Add(CsSendError, 1)
		return err
	}
	t.statistics.Add(CsSendOkay, int64(size))
	return nil
}

func (t *StreamSocket) ReadMsg() (*FrameMessage, error) {
	if HasLog(LOG) {
		Log("StreamSocket.ReadMsg: %s", t)
	}
	if !t.IsOk() {
		return nil, NewErr("%s not okay", t)
	}
	if t.message == nil { // default is stream message
		t.message = &StreamMessagerImpl{}
	}
	frame, err := t.message.Receive(t.connection, t.maxSize, t.minSize)
	if err != nil {
		return nil, err
	}
	size := len(frame.frame)
	t.statistics.Add(CsRecvOkay, int64(size))
	return frame, nil
}

type SocketClientImpl struct {
	*StreamSocket
	lock          sync.RWMutex
	listener      ClientListener
	newTime       int64
	connectedTime int64
	private       interface{}
	status        uint8
	timeout       int64 // sec for read and write timeout
	remoteAddr    string
	localAddr     string
}

func NewSocketClient(address string, message Messager) *SocketClientImpl {
	return &SocketClientImpl{
		StreamSocket: &StreamSocket{
			maxSize:    1514,
			minSize:    15,
			message:    message,
			statistics: NewSafeStrInt64(),
			out:        NewSubLogger(address),
			address:    address,
		},
		newTime: time.Now().Unix(),
		status:  ClInit,
	}
}

func (s *SocketClientImpl) Out() *SubLogger {
	if s.out == nil {
		s.out = NewSubLogger(s.address)
	}
	return s.out
}

func (s *SocketClientImpl) State() string {
	switch s.Status() {
	case ClInit:
		return "initialized"
	case ClConnected:
		return "connected"
	case ClUnAuth:
		return "unauthenticated"
	case ClAuth:
		return "authenticated"
	case ClClosed:
		return "closed"
	case ClConnecting:
		return "connecting"
	case ClTerminal:
		return "terminal"
	}
	return ""
}

func (s *SocketClientImpl) Retry() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.connection != nil ||
		s.status == ClTerminal ||
		s.status == ClUnAuth {
		return false
	}
	s.status = ClConnecting
	return true
}

func (s *SocketClientImpl) Status() uint8 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.status
}

func (s *SocketClientImpl) UpTime() int64 {
	return time.Now().Unix() - s.newTime
}

func (s *SocketClientImpl) AliveTime() int64 {
	if s.connectedTime == 0 {
		return 0
	}
	return time.Now().Unix() - s.connectedTime
}

// Get server address for client or remote address from server.
func (s *SocketClientImpl) Address() string {
	return s.address
}

func (s *SocketClientImpl) String() string {
	return s.Address()
}

func (s *SocketClientImpl) Private() interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.private
}

func (s *SocketClientImpl) SetPrivate(v interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.private = v
}

func (s *SocketClientImpl) MaxSize() int {
	return s.maxSize
}

func (s *SocketClientImpl) SetMaxSize(value int) {
	s.maxSize = value
}

func (s *SocketClientImpl) MinSize() int {
	return s.minSize
}

func (s *SocketClientImpl) Have(state uint8) bool {
	return s.Status() == state
}

func (s *SocketClientImpl) Statistics() map[string]int64 {
	sts := make(map[string]int64)
	s.statistics.Copy(sts)
	return sts
}

func (s *SocketClientImpl) SetListener(listener ClientListener) {
	s.listener = listener
}

// Get actual local address
func (s *SocketClientImpl) LocalAddr() string {
	return s.localAddr
}

// Get actual remote address
func (s *SocketClientImpl) RemoteAddr() string {
	return s.remoteAddr
}

func (s *SocketClientImpl) SetTimeout(v int64) {
	s.timeout = v
}

func (s *SocketClientImpl) updateConn(conn net.Conn) {
	if conn != nil {
		s.connection = conn
		s.connectedTime = time.Now().Unix()
		s.localAddr = conn.LocalAddr().String()
		s.remoteAddr = conn.RemoteAddr().String()
	} else {
		if s.connection != nil {
			s.connection.Close()
		}
		s.connection = nil
		s.message.Flush()
	}
}

func (s *SocketClientImpl) SetConnection(conn net.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.updateConn(conn)
	s.status = ClConnected
}

// Socket Server

const (
	SsRecv   = "recv"
	SsDeny   = "deny"
	SsAlive  = "alive"
	SsSend   = "send"
	SsDrop   = "dropped"
	SsAccept = "accept"
	SsClose  = "closed"
)

type ServerListener struct {
	OnClient func(client SocketClient) error
	OnClose  func(client SocketClient) error
	ReadAt   func(client SocketClient, f *FrameMessage) error
}

type ReadClient func(client SocketClient, f *FrameMessage) error

type SocketServer interface {
	Listen() (err error)
	Close()
	Accept()
	ListClient() <-chan SocketClient
	OffClient(client SocketClient)
	TotalClient() int
	Loop(call ServerListener)
	Read(client SocketClient, ReadAt ReadClient)
	String() string
	Address() string
	Statistics() map[string]int64
	SetTimeout(v int64)
}

// TODO keepalive to release zombie connections.
type SocketServerImpl struct {
	lock       sync.RWMutex
	statistics *SafeStrInt64
	address    string
	maxClient  int
	clients    *SafeStrMap
	onClients  chan SocketClient
	offClients chan SocketClient
	close      func()
	timeout    int64 // sec for read and write timeout
	WrQus      int   // per frames.
}

func NewSocketServer(listen string) *SocketServerImpl {
	return &SocketServerImpl{
		address:    listen,
		statistics: NewSafeStrInt64(),
		maxClient:  128,
		clients:    NewSafeStrMap(1024),
		onClients:  make(chan SocketClient, 1024),
		offClients: make(chan SocketClient, 1024),
		WrQus:      1024,
	}
}

func (t *SocketServerImpl) ListClient() <-chan SocketClient {
	list := make(chan SocketClient, 32)
	Go(func() {
		t.clients.Iter(func(k string, v interface{}) {
			if client, ok := v.(SocketClient); ok {
				list <- client
			}
		})
		list <- nil
	})
	return list
}

func (t *SocketServerImpl) TotalClient() int {
	return t.clients.Len()
}

func (t *SocketServerImpl) OffClient(client SocketClient) {
	Warn("SocketServerImpl.OffClient %s", client)
	if client != nil {
		t.offClients <- client
	}
}

func (t *SocketServerImpl) doOnClient(call ServerListener, client SocketClient) {
	Info("SocketServerImpl.doOnClient: %s ?", client)
	_ = t.clients.Set(client.RemoteAddr(), client)
	if call.OnClient != nil {
		_ = call.OnClient(client)
		if call.ReadAt != nil {
			Go(func() { t.Read(client, call.ReadAt) })
		}
	}
}

func (t *SocketServerImpl) doOffClient(call ServerListener, client SocketClient) {
	Info("SocketServerImpl.doOffClient: %s ?", client)
	addr := client.RemoteAddr()
	if _, ok := t.clients.GetEx(addr); ok {
		Info("SocketServerImpl.doOffClient: close %s", addr)
		t.statistics.Add(SsClose, 1)
		if call.OnClose != nil {
			_ = call.OnClose(client)
		}
		client.Close()
		t.clients.Del(addr)
		t.statistics.Add(SsAlive, -1)
	}
}

func (t *SocketServerImpl) Loop(call ServerListener) {
	Debug("SocketServerImpl.Loop")
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

func (t *SocketServerImpl) Read(client SocketClient, ReadAt ReadClient) {
	Log("SocketServerImpl.Read: %s", client)
	done := make(chan bool, 2)
	queue := make(chan *FrameMessage, t.WrQus)
	Go(func() {
		for {
			select {
			case frame := <-queue:
				if err := ReadAt(client, frame); err != nil {
					Error("SocketServerImpl.Read: readAt %s", err)
					return
				}
			case <-done:
				return
			}
		}
	})
	for {
		frame, err := client.ReadMsg()
		if err != nil || frame.size <= 0 {
			if frame != nil {
				Error("SocketServerImpl.Read: %s %d", client, frame.size)
			} else {
				Error("SocketServerImpl.Read: %s %s", client, err)
			}
			done <- true
			t.OffClient(client)
			break
		}
		t.statistics.Add(SsRecv, 1)
		if HasLog(LOG) {
			Log("SocketServerImpl.Read: length: %d ", frame.size)
			Log("SocketServerImpl.Read: frame : %x", frame)
		}
		queue <- frame
	}
}

func (t *SocketServerImpl) Close() {
	if t.close != nil {
		t.close()
	}
}

func (t *SocketServerImpl) Address() string {
	return t.address
}

func (t *SocketServerImpl) String() string {
	return t.Address()
}

func (t *SocketServerImpl) Statistics() map[string]int64 {
	sts := make(map[string]int64, 32)
	t.statistics.Copy(sts)
	return sts
}

func (t *SocketServerImpl) SetTimeout(v int64) {
	t.timeout = v
}

// Previous process when accept connection,
// and allowed accept new connection, will return true.
func (t *SocketServerImpl) preAccept(conn net.Conn) bool {
	addr := conn.RemoteAddr()
	Debug("SocketServerImpl.preAccept: %s", addr)
	t.statistics.Add(SsAccept, 1)
	alive := t.statistics.Get(SsAlive)
	if alive >= int64(t.maxClient) {
		Debug("SocketServerImpl.preAccept: close %s", addr)
		t.statistics.Add(SsDeny, 1)
		t.statistics.Add(SsClose, 1)
		conn.Close()
		return false
	}
	Debug("SocketServerImpl.preAccept: allow %s", addr)
	t.statistics.Add(SsAlive, 1)
	return true
}
