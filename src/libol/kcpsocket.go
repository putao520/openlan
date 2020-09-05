package libol

import (
	"github.com/xtaci/kcp-go/v5"
	"net"
	"time"
)

type KcpConfig struct {
	Block        kcp.BlockCrypt
	DataShards   int           // default 1024
	ParityShards int           // default 3
	Timeout      time.Duration // ns
}

var defaultKcpConfig = KcpConfig{
	Block:        nil,
	DataShards:   1024,
	ParityShards: 3,
	Timeout:      120 * time.Second,
}

type KcpServer struct {
	*SocketServerImpl
	kcpCfg   *KcpConfig
	listener *kcp.Listener
}

func NewKcpServer(listen string, cfg *KcpConfig) *KcpServer {
	if cfg == nil {
		cfg = &defaultKcpConfig
	}
	k := &KcpServer{
		kcpCfg:           cfg,
		SocketServerImpl: NewSocketServer(listen),
	}
	k.close = k.Close
	return k
}

func (k *KcpServer) Listen() (err error) {
	k.listener, err = kcp.ListenWithOptions(
		k.address,
		k.kcpCfg.Block,
		k.kcpCfg.DataShards,
		k.kcpCfg.ParityShards)
	if err != nil {
		k.listener = nil
		return err
	}
	Info("KcpServer.Listen: kcp://%s", k.address)
	return nil
}

func (k *KcpServer) Close() {
	if k.listener != nil {
		_ = k.listener.Close()
		Info("KcpServer.Close: %s", k.address)
		k.listener = nil
	}
}

func (k *KcpServer) Accept() {
	Debug("KcpServer.Accept")
	promise := Promise{
		First:  2 * time.Second,
		MinInt: 5 * time.Second,
		MaxInt: 30 * time.Second,
	}
	promise.Done(func() error {
		if err := k.Listen(); err != nil {
			Warn("KcpServer.Accept: %s", err)
			return err
		}
		return nil
	})
	defer k.Close()
	for {
		conn, err := k.listener.AcceptKCP()
		if err != nil {
			Error("KcpServer.Accept: %s", err)
			continue
		}
		if !k.preAccept(conn) {
			continue
		}
		conn.SetStreamMode(true)
		conn.SetWriteDelay(false)
		conn.SetACKNoDelay(false)
		k.onClients <- NewKcpClientFromConn(conn, k.kcpCfg)
	}
}

// Client Implement

type KcpClient struct {
	*SocketClientImpl
	kcpCfg *KcpConfig
}

func NewKcpClient(addr string, cfg *KcpConfig) *KcpClient {
	if cfg == nil {
		cfg = &defaultKcpConfig
	}
	c := &KcpClient{
		kcpCfg: cfg,
		SocketClientImpl: NewSocketClient(addr, &StreamMessagerImpl{
			timeout: cfg.Timeout,
		}),
	}
	c.connector = c.Connect
	return c
}

func NewKcpClientFromConn(conn net.Conn, cfg *KcpConfig) *KcpClient {
	if cfg == nil {
		cfg = &defaultKcpConfig
	}
	addr := conn.RemoteAddr().String()
	c := &KcpClient{
		SocketClientImpl: NewSocketClient(addr, &StreamMessagerImpl{
			timeout: cfg.Timeout,
		}),
	}
	c.updateConn(conn)
	c.connector = c.Connect
	return c
}

func (c *KcpClient) Connect() error {
	if !c.Retry() {
		return nil
	}
	c.out.Info("KcpClient.Connect: kcp://%s", c.address)
	conn, err := kcp.DialWithOptions(
		c.address,
		c.kcpCfg.Block,
		c.kcpCfg.DataShards,
		c.kcpCfg.DataShards)
	if err != nil {
		return err
	}
	conn.SetStreamMode(true)
	conn.SetWriteDelay(false)
	conn.SetACKNoDelay(false)
	c.SetConnection(conn)
	if c.listener.OnConnected != nil {
		_ = c.listener.OnConnected(c)
	}
	return nil
}

func (c *KcpClient) Close() {
	c.out.Info("KcpClient.Close: %v", c.IsOk())
	c.lock.Lock()
	if c.connection != nil {
		if c.status != ClTerminal {
			c.status = ClClosed
		}
		c.out.Info("KcpClient.Close")
		_ = c.connection.Close()
		c.connection = nil
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
	c.SetStatus(ClTerminal)
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
