package libol

import (
	"net"
	"time"
)

type UdpConfig struct {
	Key string
}

var defaultUdpConfig = UdpConfig{
	Key: "78fxojvnu",
}

type UdpServer struct {
	socketServer
	udpCfg  *UdpConfig
	listener net.Listener
}

func NewUdpServer(listen string, cfg *UdpConfig) *UdpServer {
	k := &UdpServer{
		socketServer: socketServer{
			addr:       listen,
			sts:        ServerSts{},
			maxClient:  1024,
			clients:    make(map[SocketClient]bool, 1024),
			onClients:  make(chan SocketClient, 4),
			offClients: make(chan SocketClient, 8),
		},
	}
	if k.udpCfg == nil {
		k.udpCfg = &defaultUdpConfig
	}
	k.close = k.Close
	if err := k.Listen(); err != nil {
		Debug("NewUdpServer: %s", err)
	}
	return k
}

func (k *UdpServer) Listen() (err error) {
	k.listener, err = XDPListen(k.addr)
	if err != nil {
		k.listener = nil
		return err
	}
	Info("UdpServer.Listen: udp://%s", k.addr)
	return nil
}

func (k *UdpServer) Close() {
	if k.listener != nil {
		_ = k.listener.Close()
		Info("UdpServer.Close: %s", k.addr)
		k.listener = nil
	}
}

func (k *UdpServer) Accept() {
	for {
		if k.listener != nil {
			break
		}
		if err := k.Listen(); err != nil {
			Warn("UdpServer.Accept: %s", err)
		}
		time.Sleep(time.Second * 5)
	}
	defer k.Close()
	for {
		conn, err := k.listener.Accept()
		if err != nil {
			Error("TcpServer.Accept: %s", err)
			return
		}
		k.sts.AcpCount++
		k.onClients <- NewUdpClientFromConn(conn)
	}
}

// Client Implement

type UdpClient struct {
	socketClient
	udpCfg *UdpConfig
}

func NewUdpClient(addr string, cfg *UdpConfig) *UdpClient {
	c := &UdpClient{
		udpCfg: cfg,
		socketClient: socketClient{
			addr:    addr,
			NewTime: time.Now().Unix(),
			dataStream: dataStream{
				maxSize: 1514,
				minSize: 15,
				message: &DataGramMessage{},
			},
			status: CL_INIT,
		},
	}
	c.connect = c.Connect
	if c.udpCfg == nil {
		c.udpCfg = &defaultUdpConfig
	}
	return c
}

func NewUdpClientFromConn(conn net.Conn) *UdpClient {
	c := &UdpClient{
		socketClient: socketClient{
			addr: conn.RemoteAddr().String(),
			dataStream: dataStream{
				conn:    conn,
				maxSize: 1514,
				minSize: 15,
				message: &DataGramMessage{},
			},
			NewTime: time.Now().Unix(),
		},
	}
	c.connect = c.Connect
	return c
}

func (c *UdpClient) LocalAddr() string {
	if c.conn != nil {
		return c.conn.LocalAddr().String()
	}
	return c.addr
}

func (c *UdpClient) Connect() error {
	c.lock.Lock()
	if c.conn != nil || c.status == CL_TERMINAL || c.status == CL_UNAUTH {
		c.lock.Unlock()
		return nil
	}
	c.status = CL_CONNECTING
	c.lock.Unlock()

	Info("UdpClient.Connect: udp://%s", c.addr)
	conn, err := net.Dial("udp", c.addr)
	if err == nil {
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

func (c *UdpClient) Close() {
	c.lock.Lock()
	if c.conn != nil {
		if c.status != CL_TERMINAL {
			c.status = CL_CLOSED
		}
		Info("UdpClient.Close: %s", c.addr)
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

func (c *UdpClient) Terminal() {
	c.SetStatus(CL_TERMINAL)
	c.Close()
}

func (c *UdpClient) SetStatus(v uint8) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status != v {
		if c.listener.OnStatus != nil {
			c.listener.OnStatus(c, c.status, v)
		}
		c.status = v
	}
}
