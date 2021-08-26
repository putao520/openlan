package libol

import (
	"github.com/xtaci/kcp-go/v5"
	"net"
	"time"
)

type UdpConfig struct {
	Block   kcp.BlockCrypt
	Timeout time.Duration // ns
	Clients int
	RdQus   int // per frames
	WrQus   int // per frames
}

var defaultUdpConfig = UdpConfig{
	Timeout: 120 * time.Second,
	Clients: 1024,
}

type UdpServer struct {
	*SocketServerImpl
	udpCfg   *UdpConfig
	listener net.Listener
}

func NewUdpServer(listen string, cfg *UdpConfig) *UdpServer {
	if cfg == nil {
		cfg = &defaultUdpConfig
	}
	if cfg.Clients == 0 {
		cfg.Clients = defaultUdpConfig.Clients
	}
	k := &UdpServer{
		udpCfg:           cfg,
		SocketServerImpl: NewSocketServer(listen),
	}
	k.close = k.Close
	return k
}

func (k *UdpServer) Listen() (err error) {
	k.listener, err = XDPListen(k.address, k.udpCfg.Clients, k.udpCfg.RdQus*2)
	if err != nil {
		k.listener = nil
		return err
	}
	Info("UdpServer.Listen: udp://%s", k.address)
	return nil
}

func (k *UdpServer) Close() {
	if k.listener != nil {
		_ = k.listener.Close()
		Info("UdpServer.Close: %s", k.address)
		k.listener = nil
	}
}

func (k *UdpServer) Accept() {
	promise := Promise{
		First:  2 * time.Second,
		MinInt: 5 * time.Second,
		MaxInt: 30 * time.Second,
	}
	promise.Done(func() error {
		if err := k.Listen(); err != nil {
			Warn("UdpServer.Accept: %s", err)
			return err
		}
		return nil
	})
	defer k.Close()
	for {
		conn, err := k.listener.Accept()
		if k.preAccept(conn, err) != nil {
			continue
		}
		k.onClients <- NewUdpClientFromConn(conn, k.udpCfg)
	}
}

// Client Implement

type UdpClient struct {
	*SocketClientImpl
	udpCfg *UdpConfig
}

func NewUdpClient(addr string, cfg *UdpConfig) *UdpClient {
	if cfg == nil {
		cfg = &defaultUdpConfig
	}
	c := &UdpClient{
		udpCfg: cfg,
		SocketClientImpl: NewSocketClient(addr, &PacketMessagerImpl{
			timeout: cfg.Timeout,
			block:   cfg.Block,
			bufSize: cfg.RdQus * MaxFrame,
		}),
	}
	return c
}

func NewUdpClientFromConn(conn net.Conn, cfg *UdpConfig) *UdpClient {
	if cfg == nil {
		cfg = &defaultUdpConfig
	}
	addr := conn.RemoteAddr().String()
	c := &UdpClient{
		SocketClientImpl: NewSocketClient(addr, &PacketMessagerImpl{
			timeout: cfg.Timeout,
			block:   cfg.Block,
			bufSize: cfg.RdQus * MaxFrame,
		}),
	}
	c.updateConn(conn)
	return c
}

func (c *UdpClient) Connect() error {
	if !c.Retry() {
		return nil
	}
	c.out.Info("UdpClient.Connect: udp://%s", c.address)
	conn, err := net.Dial("udp", c.address)
	if err != nil {
		return err
	}
	c.SetConnection(conn)
	if c.listener.OnConnected != nil {
		_ = c.listener.OnConnected(c)
	}
	return nil
}

func (c *UdpClient) Close() {
	c.out.Debug("UdpClient.Close: %v", c.IsOk())
	c.lock.Lock()
	if c.connection != nil {
		if c.status != ClTerminal {
			c.status = ClClosed
		}
		c.out.Info("UdpClient.Close")
		c.updateConn(nil)
		c.private = nil
		c.lock.Unlock()
		if c.listener.OnClose != nil {
			_ = c.listener.OnClose(c)
		}
	} else {
		c.lock.Unlock()
	}
}

func (c *UdpClient) Terminal() {
	c.SetStatus(ClTerminal)
	c.Close()
}

func (c *UdpClient) SetStatus(v SocketStatus) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status != v {
		if c.listener.OnStatus != nil {
			c.listener.OnStatus(c, c.status, v)
		}
		c.status = v
	}
}
