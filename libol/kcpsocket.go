package libol

import (
	"github.com/xtaci/kcp-go/v5"
	"net"
	"time"
)

type KcpConfig struct {
	block        kcp.BlockCrypt
	dataShards   int // default 1024
	parityShards int // default 3
}

var defaultKcpConfig = KcpConfig{
	block:        nil,
	dataShards:   1024,
	parityShards: 3,
}

type KcpServer struct {
	socketServer
	kcpCfg   *KcpConfig
	listener *kcp.Listener
}

func NewKcpServer(listen string, cfg *KcpConfig) *KcpServer {
	k := &KcpServer{
		kcpCfg: cfg,
		socketServer: socketServer{
			addr:       listen,
			sts:        ServerSts{},
			maxClient:  1024,
			clients:    make(map[SocketClient]bool, 1024),
			onClients:  make(chan SocketClient, 4),
			offClients: make(chan SocketClient, 8),
		},
	}
	if k.kcpCfg == nil {
		k.kcpCfg = &defaultKcpConfig
	}
	k.close = k.Close
	if err := k.Listen(); err != nil {
		Debug("NewKcpServer: %s", err)
	}
	return k
}

func (k *KcpServer) Listen() (err error) {
	k.listener, err = kcp.ListenWithOptions(k.addr, k.kcpCfg.block, k.kcpCfg.dataShards, k.kcpCfg.parityShards)
	if err != nil {
		k.listener = nil
		return err
	}
	Info("KcpServer.Listen: kcp://%s", k.addr)
	return nil
}

func (k *KcpServer) Close() {
	if k.listener != nil {
		_ = k.listener.Close()
		Info("KcpServer.Close: %s", k.addr)
		k.listener = nil
	}
}

func (k *KcpServer) Accept() {
	Debug("KcpServer.Accept")

	for {
		if k.listener != nil {
			break
		}
		if err := k.Listen(); err != nil {
			Warn("KcpServer.Accept: %s", err)
		}
		time.Sleep(time.Second * 5)
	}
	defer k.Close()
	for {
		conn, err := k.listener.AcceptKCP()
		if err != nil {
			Error("KcpServer.Accept: %s", err)
			return
		}
		k.sts.AcpCount++
		conn.SetStreamMode(true)
		conn.SetWriteDelay(false)
		conn.SetACKNoDelay(false)
		k.onClients <- NewKcpClientFromConn(conn)
	}
}

// Client Implement

type KcpClient struct {
	socketClient
	kcpCfg *KcpConfig
}

func NewKcpClient(addr string, cfg *KcpConfig) *KcpClient {
	c := &KcpClient{
		kcpCfg: cfg,
		socketClient: socketClient{
			addr:    addr,
			NewTime: time.Now().Unix(),
			dataStream: dataStream{
				maxSize: 1514,
				minSize: 15,
			},
			status: CL_INIT,
		},
	}
	c.connect = c.Connect
	if c.kcpCfg == nil {
		c.kcpCfg = &defaultKcpConfig
	}
	return c
}

func NewKcpClientFromConn(conn net.Conn) *KcpClient {
	c := &KcpClient{
		socketClient: socketClient{
			addr: conn.RemoteAddr().String(),
			dataStream: dataStream{
				conn:    conn,
				maxSize: 1514,
				minSize: 15,
			},
			NewTime: time.Now().Unix(),
		},
	}
	c.connect = c.Connect
	return c
}

func (c *KcpClient) LocalAddr() string {
	if c.conn != nil {
		return c.conn.LocalAddr().String()
	}
	return c.addr
}

func (c *KcpClient) Connect() error {
	c.lock.Lock()
	if c.conn != nil || c.status == CL_TERMINAL || c.status == CL_UNAUTH {
		c.lock.Unlock()
		return nil
	}
	c.status = CL_CONNECTING
	c.lock.Unlock()

	Info("KcpClient.Connect: kcp://%s", c.addr)
	conn, err := kcp.DialWithOptions(c.addr, c.kcpCfg.block, c.kcpCfg.dataShards, c.kcpCfg.dataShards)
	if err == nil {
		conn.SetStreamMode(true)
		conn.SetWriteDelay(false)
		conn.SetACKNoDelay(false)
		c.lock.Lock()
		c.conn = conn
		c.status = CL_CONNECTED
		c.lock.Unlock()
		if c.listener.OnConnected != nil {
			_ = c.listener.OnConnected(c)
		}
	}
	return nil
}

func (c *KcpClient) Close() {
	c.lock.Lock()
	if c.conn != nil {
		if c.status != CL_TERMINAL {
			c.status = CL_CLOSED
		}
		Info("KcpClient.Close: %s", c.addr)
		_ = c.conn.Close()
		c.conn = nil
		c.private = nil
		c.lock.Unlock()
		if c.listener.OnClose != nil {
			_ = c.listener.OnClose(c)
		}
	} else {
		c.lock.Unlock()
	}
}

func (c *KcpClient) Terminal() {
	c.SetStatus(CL_TERMINAL)
	c.Close()
}

func (c *KcpClient) SetStatus(v uint8) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status != v {
		if c.listener.OnStatus != nil {
			c.listener.OnStatus(c, c.status, v)
		}
		c.status = v
	}
}
