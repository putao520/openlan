package libol

import (
	"crypto/tls"
	"net"
	"time"
)

// Server Implement

type TcpServer struct {
	socketServer
	tlsCfg   *tls.Config
	listener net.Listener
}

func NewTcpServer(listen string, cfg *tls.Config) *TcpServer {
	t := &TcpServer{
		tlsCfg: cfg,
		socketServer: socketServer{
			address:    listen,
			sts:        ServerSts{},
			maxClient:  1024,
			clients:    NewSafeStrMap(1024),
			onClients:  make(chan SocketClient, 4),
			offClients: make(chan SocketClient, 8),
		},
	}
	t.close = t.Close
	if err := t.Listen(); err != nil {
		Debug("NewTcpServer: %s", err)
	}
	return t
}

func (t *TcpServer) Listen() (err error) {
	if t.tlsCfg != nil {
		t.listener, err = tls.Listen("tcp", t.address, t.tlsCfg)
		if err != nil {
			t.listener = nil
			return err
		}
		Info("TcpServer.Listen: tls://%s", t.address)
	} else {
		t.listener, err = net.Listen("tcp", t.address)
		if err != nil {
			t.listener = nil
			return err
		}
		Info("TcpServer.Listen: tcp://%s", t.address)
	}
	return nil
}

func (t *TcpServer) Close() {
	if t.listener != nil {
		_ = t.listener.Close()
		Info("TcpServer.Close: %s", t.address)
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
		t.sts.AcceptCount++
		t.onClients <- NewTcpClientFromConn(conn)
	}
}

// Client Implement

type TcpClient struct {
	socketClient
	tlsCfg *tls.Config
}

func NewTcpClient(addr string, cfg *tls.Config) *TcpClient {
	t := &TcpClient{
		tlsCfg: cfg,
		socketClient: socketClient{
			address: addr,
			newTime: time.Now().Unix(),
			dataStream: dataStream{
				maxSize: 1514,
				minSize: 15,
			},
			status: ClInit,
		},
	}
	t.connecter = t.Connect
	return t
}

func NewTcpClientFromConn(conn net.Conn) *TcpClient {
	t := &TcpClient{
		socketClient: socketClient{
			address: conn.RemoteAddr().String(),
			dataStream: dataStream{
				connection: conn,
				maxSize:    1514,
				minSize:    15,
			},
			newTime: time.Now().Unix(),
		},
	}
	t.connecter = t.Connect
	return t
}

func (t *TcpClient) Connect() error {
	t.lock.Lock()
	if t.connection != nil || t.status == ClTerminal || t.status == ClUnAuth {
		t.lock.Unlock()
		return nil
	}
	if t.connection != nil {
		_ = t.connection.Close()
		t.connection = nil
	}
	t.status = ClConnecting
	t.lock.Unlock()

	if t.tlsCfg != nil {
		Info("TcpClient.Connect: tls://%s", t.address)
	} else {
		Info("TcpClient.Connect: tcp://%s", t.address)
	}

	var err error
	var conn net.Conn
	if t.tlsCfg != nil {
		conn, err = tls.Dial("tcp", t.address, t.tlsCfg)
	} else {
		conn, err = net.Dial("tcp", t.address)
	}
	if err == nil {
		t.lock.Lock()
		t.connection = conn
		t.status = ClConnected
		t.lock.Unlock()
		if t.listener.OnConnected != nil {
			_ = t.listener.OnConnected(t)
		}
	}
	return err
}

func (t *TcpClient) Close() {
	t.lock.Lock()
	if t.connection != nil {
		if t.status != ClTerminal {
			t.status = ClClosed
		}
		Info("TcpClient.Close: %s", t.address)
		_ = t.connection.Close()
		t.connection = nil
		t.private = nil
		t.lock.Unlock()
		if t.listener.OnClose != nil {
			_ = t.listener.OnClose(t)
		}
	} else {
		t.lock.Unlock()
	}
}

func (t *TcpClient) Terminal() {
	t.SetStatus(ClTerminal)
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
