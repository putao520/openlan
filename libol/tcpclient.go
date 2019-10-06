package libol

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync"
	"time"
)

var (
	MAGIC = []byte{0xff, 0xff}
)

const (
	CL_INIT       = 0x00
	CL_CONNECTED  = 0x01
	CL_UNAUTH     = 0x02
	CL_AUTHED     = 0x03
	CL_CONNECTING = 0x04
	CL_TERMINAL   = 0x05
	CL_CLOSED     = 0x06
)

const (
	HSIZE = 0x04
)

type TcpClient struct {
	conn        *net.TCPConn
	maxsize     int
	minsize     int
	onConnected func(*TcpClient) error
	lock        sync.RWMutex

	TxOkay  uint64
	RxOkay  uint64
	TxError uint64
	Droped  uint64
	Status  uint8
	Addr    string
	NewTime int64
}

func NewTcpClient(addr string) (t *TcpClient) {
	t = &TcpClient{
		Addr:        addr,
		conn:        nil,
		maxsize:     1514,
		minsize:     15,
		TxOkay:      0,
		RxOkay:      0,
		TxError:     0,
		Droped:      0,
		Status:      CL_INIT,
		onConnected: nil,
		NewTime:     time.Now().Unix(),
	}

	return
}

func NewTcpClientFromConn(conn *net.TCPConn) (t *TcpClient) {
	t = &TcpClient{
		Addr:    conn.RemoteAddr().String(),
		conn:    conn,
		maxsize: 1514,
		minsize: 15,
		NewTime: time.Now().Unix(),
	}

	return
}

func (t *TcpClient) Connect() error {
	if t.conn != nil || t.GetStatus() == CL_TERMINAL || t.GetStatus() == CL_UNAUTH {
		return nil
	}

	Info("TcpClient.Connect %s", t.Addr)
	raddr, err := net.ResolveTCPAddr("tcp", t.Addr)
	if err != nil {
		return err
	}

	t.SetStatus(CL_CONNECTING)
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		t.conn = nil
		return err
	}
	t.conn = conn
	t.SetStatus(CL_CONNECTED)
	if t.onConnected != nil {
		t.onConnected(t)
	}

	return nil
}

func (t *TcpClient) OnConnected(on func(*TcpClient) error) {
	t.onConnected = on
}

func (t *TcpClient) Close() {
	if t.conn != nil {
		if t.GetStatus() != CL_TERMINAL {
			t.SetStatus(CL_CLOSED)
		}

		Info("TcpClient.Close %s", t.Addr)
		t.conn.Close()
		t.conn = nil
	}
}

func (t *TcpClient) recvn(buffer []byte) error {
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

	Debug("TcpClient.recvn %d", len(buffer))
	Debug("TcpClient.recvn Data: % x", buffer)

	return nil
}

func (t *TcpClient) sendn(buffer []byte) error {
	offset := 0
	size := len(buffer)
	left := size - offset

	Debug("TcpClient.sendn %d", size)
	Debug("TcpClient.sendn Data: % x", buffer)

	for left > 0 {
		tmp := buffer[offset:]
		Debug("TcpClient.sendn tmp %d", len(tmp))
		n, err := t.conn.Write(tmp)
		if err != nil {
			return err
		}
		offset += n
		left = size - offset
	}
	return nil
}

func (t *TcpClient) SendMsg(data []byte) error {
	if err := t.Connect(); err != nil {
		return err
	}

	buffer := make([]byte, HSIZE+len(data))
	copy(buffer[0:2], MAGIC)
	binary.BigEndian.PutUint16(buffer[2:4], uint16(len(data)))
	copy(buffer[HSIZE:], data)

	if err := t.sendn(buffer); err != nil {
		t.TxError++
		return err
	}

	t.TxOkay++

	return nil
}

func (t *TcpClient) RecvMsg(data []byte) (int, error) {
	Debug("TcpClient.RecvMsg %s", t)

	if !t.IsOk() {
		return -1, Errer("%s: connection isn't okay", t)
	}

	h := make([]byte, HSIZE)
	if err := t.recvn(h); err != nil {
		return -1, err
	}

	if !bytes.Equal(h[0:2], MAGIC) {
		return -1, Errer("%s: isn't right magic header", t)
	}

	size := binary.BigEndian.Uint16(h[2:4])
	if int(size) > t.maxsize || int(size) < t.minsize {
		return -1, Errer("%s: isn't right data size (%d)", t, size)
	}

	d := make([]byte, size)
	if err := t.recvn(d); err != nil {
		return -1, err
	}

	copy(data, d)
	t.RxOkay++

	return int(size), nil
}

func (t *TcpClient) GetMaxSize() int {
	return t.maxsize
}

func (t *TcpClient) SetMaxSize(value int) {
	t.maxsize = value
}

func (t *TcpClient) GetMinSize() int {
	return t.minsize
}

func (t *TcpClient) IsOk() bool {
	return t.conn != nil
}

func (t *TcpClient) IsTerminal() bool {
	return t.GetStatus() == CL_TERMINAL
}

func (t *TcpClient) IsInitialized() bool {
	return t.GetStatus() == CL_INIT
}

func (t *TcpClient) SendReq(action string, body string) error {
	data := EncInstReq(action, body)
	Debug("TcpClient.SendReq %d %s", len(data), data[6:])

	if err := t.SendMsg(data); err != nil {
		return err
	}
	return nil
}

func (t *TcpClient) SendResp(action string, body string) error {
	data := EncInstResp(action, body)
	Debug("TcpClient.SendResp %d %s", len(data), data[6:])

	if err := t.SendMsg(data); err != nil {
		return err
	}
	return nil
}

func (t *TcpClient) State() string {
	switch t.GetStatus() {
	case CL_INIT:
		return "initialized"
	case CL_CONNECTED:
		return "connected"
	case CL_UNAUTH:
		return "unauthenticated"
	case CL_AUTHED:
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

func (t *TcpClient) GetStatus() uint8 {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.Status
}

func (t *TcpClient) SetStatus(v uint8) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Status = v
}
