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
			address:    listen,
			sts:        ServerSts{},
			maxClient:  1024,
			clients:    NewSafeStrMap(1024),
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
	k.listener, err = kcp.ListenWithOptions(k.address, k.kcpCfg.block, k.kcpCfg.dataShards, k.kcpCfg.parityShards)
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
		k.sts.AcceptCount++
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
			address: addr,
			newTime: time.Now().Unix(),
			dataStream: dataStream{
				maxSize: 1514,
				minSize: 15,
			},
			status: ClInit,
		},
	}
	c.connecter = c.Connect
	if c.kcpCfg == nil {
		c.kcpCfg = &defaultKcpConfig
	}
	return c
}

func NewKcpClientFromConn(conn net.Conn) *KcpClient {
	c := &KcpClient{
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
	c.connecter = c.Connect
	return c
}

func (c *KcpClient) Connect() error {
	c.lock.Lock()
	if c.connection != nil || c.status == ClTerminal || c.status == ClUnAuth {
		c.lock.Unlock()
		return nil
	}
	c.status = ClConnecting
	c.lock.Unlock()

	Info("KcpClient.Connect: kcp://%s", c.address)
	conn, err := kcp.DialWithOptions(c.address, c.kcpCfg.block, c.kcpCfg.dataShards, c.kcpCfg.dataShards)
	if err == nil {
		conn.SetStreamMode(true)
		conn.SetWriteDelay(false)
		conn.SetACKNoDelay(false)
		c.lock.Lock()
		c.connection = conn
		c.status = ClConnected
		c.lock.Unlock()
		if c.listener.OnConnected != nil {
			_ = c.listener.OnConnected(c)
		}
	}
	return nil
}

func (c *KcpClient) Close() {
	c.lock.Lock()
	if c.connection != nil {
		if c.status != ClTerminal {
			c.status = ClClosed
		}
		Info("KcpClient.Close: %s", c.address)
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
