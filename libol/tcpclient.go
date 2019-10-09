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
	ClInit       = 0x00
	ClConnected  = 0x01
	ClUnauth     = 0x02
	ClAuthed     = 0x03
	ClConnecting = 0x04
	ClTerminal   = 0x05
	ClClosed     = 0x06
)

const (
	HSIZE = 0x04
)

type TcpClient struct {
	conn        *net.TCPConn
	maxSize     int
	minSize     int
	onConnected func(*TcpClient) error
	lock        sync.RWMutex

	TxOkay  uint64
	RxOkay  uint64
	TxError uint64
	Dropped  uint64
	Status  uint8
	Addr    string
	NewTime int64
}

func NewTcpClient(addr string) (t *TcpClient) {
	t = &TcpClient{
		Addr:        addr,
		conn:        nil,
		maxSize:     1514,
		minSize:     15,
		TxOkay:      0,
		RxOkay:      0,
		TxError:     0,
		Dropped:      0,
		Status:      ClInit,
		onConnected: nil,
		NewTime:     time.Now().Unix(),
	}

	return
}

func NewTcpClientFromConn(conn *net.TCPConn) (t *TcpClient) {
	t = &TcpClient{
		Addr:    conn.RemoteAddr().String(),
		conn:    conn,
		maxSize: 1514,
		minSize: 15,
		NewTime: time.Now().Unix(),
	}

	return
}

func (t *TcpClient) Connect() error {
	if t.conn != nil || t.GetStatus() == ClTerminal || t.GetStatus() == ClUnauth {
		return nil
	}

	Info("TcpClient.Connect %s", t.Addr)
	rAddr, err := net.ResolveTCPAddr("tcp", t.Addr)
	if err != nil {
		return err
	}

	t.SetStatus(ClConnecting)
	conn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		t.conn = nil
		return err
	}
	t.conn = conn
	t.SetStatus(ClConnected)
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
		if t.GetStatus() != ClTerminal {
			t.SetStatus(ClClosed)
		}

		Info("TcpClient.Close %s", t.Addr)
		t.conn.Close()
		t.conn = nil
	}
}

func (t *TcpClient) recvX(buffer []byte) error {
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

	Debug("TcpClient.recvX %d", len(buffer))
	Debug("TcpClient.recvX Data: % x", buffer)

	return nil
}

func (t *TcpClient) sendX(buffer []byte) error {
	offset := 0
	size := len(buffer)
	left := size - offset

	Debug("TcpClient.sendX %d", size)
	Debug("TcpClient.sendX Data: % x", buffer)

	for left > 0 {
		tmp := buffer[offset:]
		Debug("TcpClient.sendX tmp %d", len(tmp))
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

	if err := t.sendX(buffer); err != nil {
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
	if err := t.recvX(h); err != nil {
		return -1, err
	}

	if !bytes.Equal(h[0:2], MAGIC) {
		return -1, Errer("%s: isn't right magic header", t)
	}

	size := binary.BigEndian.Uint16(h[2:4])
	if int(size) > t.maxSize || int(size) < t.minSize {
		return -1, Errer("%s: isn't right data size (%d)", t, size)
	}

	d := make([]byte, size)
	if err := t.recvX(d); err != nil {
		return -1, err
	}

	copy(data, d)
	t.RxOkay++

	return int(size), nil
}

func (t *TcpClient) GetMaxSize() int {
	return t.maxSize
}

func (t *TcpClient) SetMaxSize(value int) {
	t.maxSize = value
}

func (t *TcpClient) GetMinSize() int {
	return t.minSize
}

func (t *TcpClient) IsOk() bool {
	return t.conn != nil
}

func (t *TcpClient) IsTerminal() bool {
	return t.GetStatus() == ClTerminal
}

func (t *TcpClient) IsInitialized() bool {
	return t.GetStatus() == ClInit
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
	case ClInit:
		return "initialized"
	case ClConnected:
		return "connected"
	case ClUnauth:
		return "unauthenticated"
	case ClAuthed:
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

func (t *TcpClient) UpTime() int64 {
	return time.Now().Unix() - t.NewTime
}

func (t *TcpClient) String() string {
	return t.Addr
}

func (t *TcpClient) Terminal() {
	t.SetStatus(ClTerminal)
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
