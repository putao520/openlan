package libol

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"net"
	"sync"
	"time"
)

type TcpServerListener struct {
	OnClient func(client *TcpClient) error
	OnClose  func(client *TcpClient) error
	ReadAt   func(client *TcpClient, p []byte) error
}

type TcpServerSts struct {
	RxCount  int64
	TxCount  int64
	DrpCount int64
	AcpCount int64
	ClsCount int64
}

type TcpServer struct {
	Addr    string
	TlsConf *tls.Config
	Sts     TcpServerSts

	listener   net.Listener
	maxClient  int
	clients    map[*TcpClient]bool
	onClients  chan *TcpClient
	offClients chan *TcpClient
}

func NewTcpServer(listen string, config *tls.Config) (t *TcpServer) {
	t = &TcpServer{
		Addr:       listen,
		TlsConf:    config,
		Sts:        TcpServerSts{},
		maxClient:  1024,
		clients:    make(map[*TcpClient]bool, 1024),
		onClients:  make(chan *TcpClient, 4),
		offClients: make(chan *TcpClient, 8),
	}

	if err := t.Listen(); err != nil {
		Debug("NewTcpServer %s", err)
	}

	return
}

func (t *TcpServer) Listen() (err error) {
	schema := "tcp"
	if t.TlsConf != nil {
		schema = "tls"
	}
	Info("TcpServer.Start %s://%s", schema, t.Addr)

	if t.TlsConf != nil {
		t.listener, err = tls.Listen("tcp", t.Addr, t.TlsConf)
		if err != nil {
			Info("TcpServer.Listen: %s", err)
			return
		}
	} else {
		t.listener, err = net.Listen("tcp", t.Addr)
		if err != nil {
			Info("TcpServer.Listen: %s", err)
			t.listener = nil
			return
		}
	}

	return nil
}

func (t *TcpServer) Close() {
	if t.listener != nil {
		t.listener.Close()
		Info("TcpServer.Close: %s", t.Addr)
		t.listener = nil
	}
}

func (t *TcpServer) Accept() {
	Debug("TcpServer.Accept")
	if t.listener == nil {
		Error("TcpServer.Accept: invalid listener")
	}

	defer t.Close()
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			Error("TcpServer.Accept: %s", err)
			return
		}

		t.Sts.AcpCount++
		t.onClients <- NewTcpClientFromConn(conn)
	}
}

func (t *TcpServer) Loop(call TcpServerListener) {
	Debug("TcpServer.Loop")
	defer t.Close()
	for {
		select {
		case client := <-t.onClients:
			Debug("TcpServer.addClient %s", client.Addr)
			t.clients[client] = true
			if call.OnClient != nil {
				call.OnClient(client)
				if call.ReadAt != nil {
					go t.Read(client, call.ReadAt)
				}
			}
		case client := <-t.offClients:
			if ok := t.clients[client]; ok {
				Debug("TcpServer.delClient %s", client.Addr)
				t.Sts.ClsCount++
				if call.OnClose != nil {
					call.OnClose(client)
				}
				client.Close()
				delete(t.clients, client)
			}
		}
	}
}

func (t *TcpServer) Read(client *TcpClient, ReadAt func(client *TcpClient, p []byte) error) {
	Debug("TcpServer.Read: %s", client.Addr)
	for {
		data := make([]byte, 4096)
		length, err := client.ReadMsg(data)
		if err != nil {
			Error("TcpServer.Read: %s", err)
			t.offClients <- client
			break
		}
		if length <= 0 {
			continue
		}
		t.Sts.RxCount++
		Debug("TcpServer.Read: length: %d ", length)
		Debug("TcpServer.Read: data  : %x", data[:length])
		if err := ReadAt(client, data[:length]); err != nil {
			Error("TcpServer.Read: do-write %s", err)
			break
		}
	}
}

type TcpClientListener struct {
	OnClose     func(client *TcpClient) error
	OnConnected func(client *TcpClient) error
	OnStatus    func(client *TcpClient, old, new uint8)
}

const (
	CL_INIT       = 0x00
	CL_CONNECTED  = 0x01
	CL_UNAUTH     = 0x02
	CL_AUEHED     = 0x03
	CL_CONNECTING = 0x04
	CL_TERMINAL   = 0x05
	CL_CLOSED     = 0x06
)

type TcpClientSts struct {
	TxOkay  uint64
	RxOkay  uint64
	TxError uint64
	Dropped uint64
}

type TcpClient struct {
	Addr     string
	NewTime  int64
	TlsConf  *tls.Config
	Sts      TcpClientSts
	Listener TcpClientListener

	conn    net.Conn
	maxSize int
	minSize int
	lock    sync.RWMutex
	status  uint8
}

func NewTcpClient(addr string, config *tls.Config) (t *TcpClient) {
	t = &TcpClient{
		Addr:    addr,
		NewTime: time.Now().Unix(),
		Sts:     TcpClientSts{},
		TlsConf: config,
		maxSize: 1514,
		minSize: 15,
		status:  CL_INIT,
	}

	return
}

func NewTcpClientFromConn(conn net.Conn) (t *TcpClient) {
	t = &TcpClient{
		Addr:    conn.RemoteAddr().String(),
		conn:    conn,
		maxSize: 1514,
		minSize: 15,
		NewTime: time.Now().Unix(),
	}

	return
}

func (t *TcpClient) LocalAddr() string {
	return t.conn.LocalAddr().String()
}

func (t *TcpClient) Connect() (err error) {
	t.lock.Lock()
	if t.conn != nil || t.status == CL_TERMINAL || t.status == CL_UNAUTH {
		t.lock.Unlock()
		return nil
	}
	schema := "tcp"
	if t.TlsConf != nil {
		schema = "tls"
	}
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
	Info("TcpClient.Connect %s://%s", schema, t.Addr)
	t.status = CL_CONNECTING
	t.lock.Unlock()

	var conn net.Conn
	if t.TlsConf != nil {
		conn, err = tls.Dial("tcp", t.Addr, t.TlsConf)
	} else {
		conn, err = net.Dial("tcp", t.Addr)
	}
	if err == nil {
		t.lock.Lock()
		t.conn = conn
		t.status = CL_CONNECTED
		t.lock.Unlock()
		if t.Listener.OnConnected != nil {
			t.Listener.OnConnected(t)
		}
	}

	return err
}

func (t *TcpClient) Close() {
	t.lock.Lock()
	if t.conn != nil {
		if t.status != CL_TERMINAL {
			t.status = CL_CLOSED
		}
		Info("TcpClient.Close %s", t.Addr)
		t.conn.Close()
		t.conn = nil
		t.lock.Unlock()

		if t.Listener.OnClose != nil {
			t.Listener.OnClose(t)
		}
	} else {
		t.lock.Unlock()
	}
}

func (t *TcpClient) ReadFull(buffer []byte) error {
	Debug("TcpClient.ReadFull %d", len(buffer))

	offset := 0
	left := len(buffer)
	for left > 0 {
		tmp := make([]byte, left)
		n, err := t.conn.Read(tmp)
		if err != nil {
			return err
		}
		copy(buffer[offset:], tmp)
		offset += n
		left -= n
	}

	Debug("TcpClient.ReadFull Data: %x", buffer)
	return nil
}

func (t *TcpClient) WriteFull(buffer []byte) error {
	offset := 0
	size := len(buffer)
	left := size - offset

	Debug("TcpClient.WriteFull %d", size)
	Debug("TcpClient.WriteFull Data: %x", buffer)

	for left > 0 {
		tmp := buffer[offset:]
		Debug("TcpClient.WriteFull tmp %d", len(tmp))
		n, err := t.conn.Write(tmp)
		if err != nil {
			return err
		}
		offset += n
		left = size - offset
	}
	return nil
}

func (t *TcpClient) WriteMsg(data []byte) error {
	if err := t.Connect(); err != nil {
		return err
	}

	size := len(data)
	buf := make([]byte, HSIZE+size)
	copy(buf[0:2], MAGIC)
	binary.BigEndian.PutUint16(buf[2:4], uint16(size))
	copy(buf[HSIZE:], data)

	if err := t.WriteFull(buf); err != nil {
		t.Sts.TxError++
		return err
	}

	t.Sts.TxOkay += uint64(size)

	return nil
}

func (t *TcpClient) ReadMsg(data []byte) (int, error) {
	Debug("TcpClient.ReadMsg %s", t)

	if !t.IsOk() {
		return -1, NewErr("%s: not okay", t)
	}

	hl := GetHeaderLen()
	buffer := make([]byte, hl+t.maxSize)
	h := buffer[:hl]
	if err := t.ReadFull(h); err != nil {
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
	if err := t.ReadFull(d); err != nil {
		return -1, err
	}

	copy(data, d)
	t.Sts.RxOkay += uint64(size)

	return len(d), nil
}

func (t *TcpClient) WriteReq(action string, body string) error {
	m := NewControlMessage(action, "= ", body)
	data := m.Encode()
	Debug("TcpClient.WriteReq %d %s", len(data), data[6:])

	if err := t.WriteMsg(data); err != nil {
		return err
	}
	return nil
}

func (t *TcpClient) WriteResp(action string, body string) error {
	m := NewControlMessage(action, ": ", body)
	data := m.Encode()
	Debug("TcpClient.WriteResp %d %s", len(data), data[6:])

	if err := t.WriteMsg(data); err != nil {
		return err
	}
	return nil
}

func (t *TcpClient) State() string {
	switch t.Status() {
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

func (t *TcpClient) UpTime() int64 {
	return time.Now().Unix() - t.NewTime
}

func (t *TcpClient) String() string {
	return t.Addr
}

func (t *TcpClient) Terminal() {
	t.SetStatus(CL_TERMINAL)
	t.Close()
}

func (t *TcpClient) Status() uint8 {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.status
}

func (t *TcpClient) SetStatus(v uint8) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.status != v {
		if t.Listener.OnStatus != nil {
			t.Listener.OnStatus(t, t.status, v)
		}
		t.status = v
	}
}

func (t *TcpClient) MaxSize() int {
	return t.maxSize
}

func (t *TcpClient) SetMaxSize(value int) {
	t.maxSize = value
}

func (t *TcpClient) MinSize() int {
	return t.minSize
}

func (t *TcpClient) IsOk() bool {
	return t.conn != nil
}

func (t *TcpClient) IsTerminal() bool {
	return t.Status() == CL_TERMINAL
}

func (t *TcpClient) IsInitialized() bool {
	return t.Status() == CL_INIT
}
