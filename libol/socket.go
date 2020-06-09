package libol

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
	ReadFull(buffer []byte) error
	WriteFull(buffer []byte) error
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
	Have(status int) bool
	Addr() string
	SetAddr(addr string)
	Sts() ClientSts
	SetListener(listener ClientListener)
}

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
	CloseClient(client SocketClient)
	Loop(call ServerListener)
	Read(client SocketClient, ReadAt ReadClient)
	String() string
	Addr() string
	Sts() ServerSts
}
