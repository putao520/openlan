package libol

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"net"
	"time"
)

// Server Implement

type TcpServer struct {
	socketServer
	tlsCfg   *tls.Config
	listener net.Listener
}

func NewTcpServer(listen string, config *tls.Config) *TcpServer {
	t := &TcpServer{
		tlsCfg: config,
		socketServer: socketServer{
			addr:       listen,
			sts:        ServerSts{},
			maxClient:  1024,
			clients:    make(map[SocketClient]bool, 1024),
			onClients:  make(chan SocketClient, 4),
			offClients: make(chan SocketClient, 8),
		},
	}
	if err := t.Listen(); err != nil {
		Debug("NewTcpServer: %s", err)
	}
	return t
}

func (t *TcpServer) Listen() (err error) {
	if t.tlsCfg != nil {
		t.listener, err = tls.Listen("tcp", t.addr, t.tlsCfg)
		if err != nil {
			return err
		}
		Info("TcpServer.Listen: tls://%s", t.addr)
	} else {
		t.listener, err = net.Listen("tcp", t.addr)
		if err != nil {
			t.listener = nil
			return err
		}
		Info("TcpServer.Listen: tcp://%s", t.addr)
	}
	return nil
}

func (t *TcpServer) Close() {
	if t.listener != nil {
		_ = t.listener.Close()
		Info("TcpServer.Close: %s", t.addr)
		t.listener = nil
	}
}

func (t *TcpServer) Accept() {
	Debug("TcpServer.Accept")

	for {
		if t.listener != nil {
			break
		}
		if err := t.Listen(); err != nil {
			Warn("TcpServer.Accept: %s", err)
		}
		time.Sleep(time.Second * 5)
	}
	defer t.Close()
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			Error("TcpServer.Accept: %s", err)
			return
		}
		t.sts.AcpCount++
		t.onClients <- NewTcpClientFromConn(conn)
	}
}

func (t *TcpServer) OffClient(client SocketClient) {
	Debug("TcpServer.OffClient %s", client)
	if client != nil {
		t.offClients <- client
	}
}

func (t *TcpServer) doOnClient(call ServerListener, client SocketClient) {
	Debug("TcpServer.doOnClient: %s", client.Addr())
	t.clients[client] = true
	if call.OnClient != nil {
		_ = call.OnClient(client)
		if call.ReadAt != nil {
			go t.Read(client, call.ReadAt)
		}
	}
}

func (t *TcpServer) doOffClient(call ServerListener, client SocketClient) {
	Debug("TcpServer.doOffClient: %s", client.Addr())
	if ok := t.clients[client]; ok {
		t.sts.ClsCount++
		if call.OnClose != nil {
			_ = call.OnClose(client)
		}
		client.Close()
		delete(t.clients, client)
	}
}

func (t *TcpServer) Loop(call ServerListener) {
	Debug("TcpServer.Loop")
	defer t.Close()
	for {
		select {
		case client := <-t.onClients:
			t.doOnClient(call, client)
		case client := <-t.offClients:
			t.doOffClient(call, client)
		}
	}
}

func (t *TcpServer) Read(client SocketClient, ReadAt ReadClient) {
	Log("TcpServer.Read: %s", client.Addr())
	data := make([]byte, MAXBUF)
	for {
		length, err := client.ReadMsg(data)
		if err != nil {
			Error("TcpServer.Read: %s", err)
			t.OffClient(client)
			break
		}
		if length <= 0 {
			continue
		}
		t.sts.RxCount++
		Log("TcpServer.Read: length: %d ", length)
		Log("TcpServer.Read: data  : %x", data[:length])
		if err := ReadAt(client, data[:length]); err != nil {
			Error("TcpServer.Read: do-write %s", err)
			break
		}
	}
}

// Client Implement

type TcpClient struct {
	socketClient
	tlsCfg *tls.Config
	conn   net.Conn
}

func NewTcpClient(addr string, config *tls.Config) *TcpClient {
	return &TcpClient{
		tlsCfg: config,
		socketClient: socketClient{
			addr:    addr,
			NewTime: time.Now().Unix(),
			sts:     ClientSts{},
			maxSize: 1514,
			minSize: 15,
			status:  CL_INIT,
		},
	}
}

func NewTcpClientFromConn(conn net.Conn) *TcpClient {
	return &TcpClient{
		conn: conn,
		socketClient: socketClient{
			addr:    conn.RemoteAddr().String(),
			maxSize: 1514,
			minSize: 15,
			NewTime: time.Now().Unix(),
		},
	}
}

func (t *TcpClient) LocalAddr() string {
	if t.conn != nil {
		return t.conn.LocalAddr().String()
	}
	return t.addr
}

func (t *TcpClient) Connect() (err error) {
	t.lock.Lock()
	if t.conn != nil || t.status == CL_TERMINAL || t.status == CL_UNAUTH {
		t.lock.Unlock()
		return nil
	}
	schema := "tcp"
	if t.tlsCfg != nil {
		schema = "tls"
	}
	if t.conn != nil {
		_ = t.conn.Close()
		t.conn = nil
	}
	Info("TcpClient.Connect: %s://%s", schema, t.addr)
	t.status = CL_CONNECTING
	t.lock.Unlock()

	var conn net.Conn
	if t.tlsCfg != nil {
		conn, err = tls.Dial("tcp", t.addr, t.tlsCfg)
	} else {
		conn, err = net.Dial("tcp", t.addr)
	}
	if err == nil {
		t.lock.Lock()
		t.conn = conn
		t.status = CL_CONNECTED
		t.lock.Unlock()
		if t.listener.OnConnected != nil {
			_ = t.listener.OnConnected(t)
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
		Info("TcpClient.Close: %s", t.addr)
		_ = t.conn.Close()
		t.conn = nil
		t.private = nil
		t.lock.Unlock()

		if t.listener.OnClose != nil {
			_ = t.listener.OnClose(t)
		}
	} else {
		t.lock.Unlock()
	}
}

func (t *TcpClient) ReadFull(buffer []byte) error {
	Log("TcpClient.ReadFull: %d", len(buffer))

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

	Log("TcpClient.ReadFull: Data %x", buffer)
	return nil
}

func (t *TcpClient) WriteFull(buffer []byte) error {
	offset := 0
	size := len(buffer)
	left := size - offset

	Log("TcpClient.WriteFull: %d", size)
	Log("TcpClient.WriteFull: Data %x", buffer)

	for left > 0 {
		tmp := buffer[offset:]
		Log("TcpClient.WriteFull: tmp %d", len(tmp))
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
		t.sts.Dropped++
		return err
	}

	size := len(data)
	buf := make([]byte, HSIZE+size)
	copy(buf[0:2], MAGIC)
	binary.BigEndian.PutUint16(buf[2:4], uint16(size))
	copy(buf[HSIZE:], data)

	if err := t.WriteFull(buf); err != nil {
		t.sts.TxError++
		return err
	}
	t.sts.TxOkay += uint64(size)

	return nil
}

func (t *TcpClient) ReadMsg(data []byte) (int, error) {
	Log("TcpClient.ReadMsg: %s", t)

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
	t.sts.RxOkay += uint64(size)

	return len(d), nil
}

func (t *TcpClient) WriteReq(action string, body string) error {
	m := NewControlMessage(action, "= ", body)
	data := m.Encode()
	Log("TcpClient.WriteReq: %d %s", len(data), data[6:])

	if err := t.WriteMsg(data); err != nil {
		return err
	}
	return nil
}

func (t *TcpClient) WriteResp(action string, body string) error {
	m := NewControlMessage(action, ": ", body)
	data := m.Encode()
	Log("TcpClient.WriteResp: %d %s", len(data), data[6:])

	if err := t.WriteMsg(data); err != nil {
		return err
	}
	return nil
}

func (t *TcpClient) Terminal() {
	t.SetStatus(CL_TERMINAL)
	t.Close()
}

func (t *TcpClient) SetStatus(v uint8) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.status != v {
		if t.listener.OnStatus != nil {
			t.listener.OnStatus(t, t.status, v)
		}
		t.status = v
	}
}

func (t *TcpClient) IsOk() bool {
	return t.conn != nil
}
