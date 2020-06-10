package libol

import (
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
	SendOkay  uint64
	RecvOkay  uint64
	SendError uint64
	Dropped   uint64
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

func (t *dataStream) WriteMsg(data []byte) error {
	if err := t.connecter(); err != nil {
		t.sts.Dropped++
		return err
	}
	if t.message == nil { // default is stream message
		t.message = &StreamMessage{}
	}
	n, err := t.message.Send(t.connection, data)
	if err != nil {
		t.sts.SendError++
		return err
	}
	t.sts.SendOkay += uint64(n)
	return nil
}

func (t *dataStream) ReadMsg(data []byte) (int, error) {
	Log("dataStream.ReadMsg: %s", t)
	if !t.IsOk() {
		return -1, NewErr("%s: not okay", t)
	}
	if t.message == nil { // default is stream message
		t.message = &StreamMessage{}
	}
	size, err := t.message.Receive(t.connection, data, t.maxSize, t.minSize)
	if err != nil {
		return size, err
	}
	t.sts.RecvOkay += uint64(size)

	return size, nil
}

func (t *dataStream) WriteReq(action string, body string) error {
	m := NewControlMessage(action, "= ", body)
	data := m.Encode()
	Log("dataStream.WriteReq: %d %s", len(data), data[6:])
	return t.WriteMsg(data)
}

func (t *dataStream) WriteResp(action string, body string) error {
	m := NewControlMessage(action, ": ", body)
	data := m.Encode()
	Log("dataStream.WriteResp: %d %s", len(data), data[6:])
	return t.WriteMsg(data)
}

type socketClient struct {
	dataStream
	lock     sync.RWMutex
	listener ClientListener
	address  string
	newTime  int64
	private  interface{}
	status   uint8
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

// Socket Server

type ServerSts struct {
	RecvCount   int64
	SendCount   int64
	DropCount   int64
	AcceptCount int64
	CloseCount  int64
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
	address    string
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
		t.sts.CloseCount++
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
		t.sts.RecvCount++
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
	return t.address
}

func (t *socketServer) String() string {
	return t.Addr()
}

func (t *socketServer) Sts() ServerSts {
	return t.sts
}
