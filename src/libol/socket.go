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

type ClientSts struct {
	SendOkay  uint64 `json:"send"`
	RecvOkay  uint64 `json:"recv"`
	SendError uint64 `json:"error"`
	Dropped   uint64 `json:"dropped"`
}

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
	SetTimeout(v int64)
}

type dataStream struct {
	message    Messager
	connection net.Conn
	sts        ClientSts
	maxSize    int
	minSize    int
	connecter  func() error
}

func (t *dataStream) String() string {
	if t.connection != nil {
		return t.connection.RemoteAddr().String()
	}
	return "unknown"
}

func (t *dataStream) IsOk() bool {
	return t.connection != nil
}

func (t *dataStream) WriteMsg(frame *FrameMessage) error {
	if err := t.connecter(); err != nil {
		t.sts.Dropped++
		return err
	}
	if t.message == nil { // default is stream message
		t.message = &StreamMessage{}
	}
	n, err := t.message.Send(t.connection, frame)
	if err != nil {
		t.sts.SendError++
		return err
	}
	t.sts.SendOkay += uint64(n)
	return nil
}

func (t *dataStream) ReadMsg() (*FrameMessage, error) {
	if HasLog(LOG) {
		Log("dataStream.ReadMsg: %s", t)
	}
	if !t.IsOk() {
		return nil, NewErr("%s: not okay", t)
	}
	if t.message == nil { // default is stream message
		t.message = &StreamMessage{}
	}
	frame, err := t.message.Receive(t.connection, t.maxSize, t.minSize)
	if err != nil {
		return nil, err
	}
	t.sts.RecvOkay += uint64(len(frame.frame))

	return frame, nil
}

func (t *dataStream) WriteReq(action string, body string) error {
	m := NewControlMessage(action, "= ", body)
	frame := m.Encode()
	Cmd("dataStream.WriteReq: %s", frame.frame)
	return t.WriteMsg(frame)
}

func (t *dataStream) WriteResp(action string, body string) error {
	m := NewControlMessage(action, ": ", body)
	frame := m.Encode()
	Cmd("dataStream.WriteRsp: %s", frame.frame)
	return t.WriteMsg(frame)
}

type socketClient struct {
	dataStream
	lock     sync.RWMutex
	listener ClientListener
	address  string
	newTime  int64
	private  interface{}
	status   uint8
	timeout  int64 // sec for read and write timeout
}

func (s *socketClient) State() string {
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

func (s *socketClient) retry() bool {
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

func (s *socketClient) Status() uint8 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.status
}

func (s *socketClient) UpTime() int64 {
	return time.Now().Unix() - s.newTime
}

func (s *socketClient) Addr() string {
	return s.address
}

func (s *socketClient) SetAddr(addr string) {
	s.address = addr
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

func (s *socketClient) LocalAddr() string {
	if s.connection != nil {
		return s.connection.LocalAddr().String()
	}
	return ""
}

func (s *socketClient) RemoteAddr() string {
	if s.connection != nil {
		return s.connection.RemoteAddr().String()
	}
	return ""
}

func (s *socketClient) SetTimeout(v int64) {
	s.timeout = v
}

// Socket Server

type ServerSts struct {
	RecvCount   int64 `json:"recv"`
	SendCount   int64 `json:"send"`
	DropCount   int64 `json:"dropped"`
	AcceptCount int64 `json:"accept"`
	CloseCount  int64 `json:"closed"`
}

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
	Loop(call ServerListener)
	Read(client SocketClient, ReadAt ReadClient)
	String() string
	Addr() string
	Sts() ServerSts
	SetTimeout(v int64)
}

type socketServer struct {
	lock       sync.RWMutex
	sts        ServerSts
	address    string
	maxClient  int
	clients    *SafeStrMap
	onClients  chan SocketClient
	offClients chan SocketClient
	close      func()
	timeout    int64 // sec for read and write timeout
}

func (t *socketServer) ListClient() <-chan SocketClient {
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

func (t *socketServer) OffClient(client SocketClient) {
	Warn("socketServer.OffClient %s", client)
	if client != nil {
		t.offClients <- client
	}
}

func (t *socketServer) doOnClient(call ServerListener, client SocketClient) {
	Info("socketServer.doOnClient: %s", client)
	_ = t.clients.Set(client.RemoteAddr(), client)
	if call.OnClient != nil {
		_ = call.OnClient(client)
		if call.ReadAt != nil {
			Go(func() { t.Read(client, call.ReadAt) })
		}
	}
}

func (t *socketServer) doOffClient(call ServerListener, client SocketClient) {
	Info("socketServer.doOffClient: %s", client)
	if _, ok := t.clients.GetEx(client.RemoteAddr()); ok {
		t.sts.CloseCount++
		if call.OnClose != nil {
			_ = call.OnClose(client)
		}
		client.Close()
		t.clients.Del(client.RemoteAddr())
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
	Log("socketServer.Read: %s", client)
	for {
		frame, err := client.ReadMsg()
		if err != nil || frame.size <= 0 {
			if frame != nil {
				Error("socketServer.Read: %s %d", client, frame.size)
			} else {
				Error("socketServer.Read: %s %s", client, err)
			}
			t.OffClient(client)
			break
		}
		t.sts.RecvCount++
		if HasLog(LOG) {
			Log("socketServer.Read: length: %d ", frame.size)
			Log("socketServer.Read: frame : %x", frame)
		}
		if err := ReadAt(client, frame); err != nil {
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
	return t.address
}

func (t *socketServer) String() string {
	return t.Addr()
}

func (t *socketServer) Sts() ServerSts {
	return t.sts
}

func (t *socketServer) SetTimeout(v int64) {
	t.timeout = v
}
