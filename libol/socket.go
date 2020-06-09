package libol

import (
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
	Connect() (err error)
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

type socketClient struct {
	lock     sync.RWMutex
	listener ClientListener
	addr     string
	NewTime  int64
	sts      ClientSts
	private  interface{}
	maxSize  int
	minSize  int
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
