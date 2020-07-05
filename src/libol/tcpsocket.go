package libol

import (
	"crypto/tls"
	"github.com/xtaci/kcp-go/v5"
	"net"
	"time"
)

type TcpConfig struct {
	Tls     *tls.Config
	Block   kcp.BlockCrypt
	Timeout time.Duration // ns
}

// Server Implement

type TcpServer struct {
	*socketServer
	tcpCfg   *TcpConfig
	listener net.Listener
}

func NewTcpServer(listen string, cfg *TcpConfig) *TcpServer {
	t := &TcpServer{
		tcpCfg:       cfg,
		socketServer: NewSocketServer(listen),
	}
	t.close = t.Close
	if err := t.Listen(); err != nil {
		Debug("NewTcpServer: %s", err)
	}
	return t
}

func (t *TcpServer) Listen() (err error) {
	if t.tcpCfg.Tls != nil {
		t.listener, err = tls.Listen("tcp", t.address, t.tcpCfg.Tls)
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
		Info("TcpServer.Accept: %s", conn.RemoteAddr())
		t.sts.AcceptCount++
		t.onClients <- NewTcpClientFromConn(conn, t.tcpCfg)
	}
}

// Client Implement

type TcpClient struct {
	socketClient
	tcpCfg *TcpConfig
}

func NewTcpClient(addr string, cfg *TcpConfig) *TcpClient {
	t := &TcpClient{
		tcpCfg: cfg,
		socketClient: socketClient{
			address: addr,
			newTime: time.Now().Unix(),
			dataStream: dataStream{
				maxSize: 1514,
				minSize: 15,
				message: &StreamMessage{
					block: cfg.Block,
				},
			},
			status: ClInit,
		},
	}
	t.connector = t.Connect
	return t
}

func NewTcpClientFromConn(conn net.Conn, cfg *TcpConfig) *TcpClient {
	t := &TcpClient{
		tcpCfg: cfg,
		socketClient: socketClient{
			address: conn.RemoteAddr().String(),
			dataStream: dataStream{
				connection: conn,
				maxSize:    1514,
				minSize:    15,
				message: &StreamMessage{
					block: cfg.Block,
				},
			},
			newTime:       time.Now().Unix(),
			connectedTime: time.Now().Unix(),
		},
	}
	t.connector = t.Connect
	return t
}

func (t *TcpClient) Connect() error {
	if !t.retry() {
		return nil
	}
	var err error
	var conn net.Conn
	if t.tcpCfg.Tls != nil {
		Info("TcpClient.Connect: tls://%s", t.address)
		conn, err = tls.Dial("tcp", t.address, t.tcpCfg.Tls)
	} else {
		Info("TcpClient.Connect: tcp://%s", t.address)
		conn, err = net.Dial("tcp", t.address)
	}
	if err != nil {
		return err
	}
	t.SetConnection(conn)
	if t.listener.OnConnected != nil {
		_ = t.listener.OnConnected(t)
	}
	return nil
}

func (t *TcpClient) Close() {
	Info("TcpClient.Close: %s %v", t.address, t.IsOk())
	t.lock.Lock()
	if t.connection != nil {
		if t.status != ClTerminal {
			t.status = ClClosed
		}
		_ = t.connection.Close()
		t.connection = nil
		t.private = nil
		t.lock.Unlock()
		if t.listener.OnClose != nil {
			_ = t.listener.OnClose(t)
		}
		Info("TcpClient.Close: %s %d", t.address, t.status)
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
