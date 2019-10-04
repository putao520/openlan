package vswitch

import (
	"net"

	"github.com/lightstar-dev/openlan-go/libol"
)

type TcpServer struct {
	Addr string

	listener   *net.TCPListener
	maxClient  int
	clients    map[*libol.TcpClient]bool
	onClients  chan *libol.TcpClient
	offClients chan *libol.TcpClient
}

func NewTcpServer(c *Config) (t *TcpServer) {
	t = &TcpServer{
		Addr:       c.TcpListen,
		listener:   nil,
		maxClient:  1024,
		clients:    make(map[*libol.TcpClient]bool, 1024),
		onClients:  make(chan *libol.TcpClient, 4),
		offClients: make(chan *libol.TcpClient, 8),
	}

	if err := t.Listen(); err != nil {
		libol.Debug("NewTcpServer %s\n", err)
	}

	return
}

func (t *TcpServer) Listen() error {
	libol.Debug("TcpServer.Start %s\n", t.Addr)

	laddr, err := net.ResolveTCPAddr("tcp", t.Addr)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		libol.Info("TcpServer.Listen: %s", err)
		t.listener = nil
		return err
	}
	t.listener = listener
	return nil
}

func (t *TcpServer) Close() {
	if t.listener != nil {
		t.listener.Close()
		libol.Info("TcpServer.Close: %s", t.Addr)
		t.listener = nil
	}
}

func (t *TcpServer) GoAccept() {
	libol.Debug("TcpServer.GoAccept")
	if t.listener == nil {
		libol.Error("TcpServer.GoAccept: invalid listener")
	}

	defer t.Close()
	for {
		conn, err := t.listener.AcceptTCP()
		if err != nil {
			libol.Error("TcpServer.GoAccept: %s", err)
			return
		}

		t.onClients <- libol.NewTcpClientFromConn(conn)
	}

	return
}

func (t *TcpServer) GoLoop(onClient func(*libol.TcpClient) error,
	onRecv func(*libol.TcpClient, []byte) error,
	onClose func(*libol.TcpClient) error) {
	libol.Debug("TcpServer.GoLoop")
	defer t.Close()
	for {
		select {
		case client := <-t.onClients:
			libol.Debug("TcpServer.addClient %s", client.Addr)
			if onClient != nil {
				onClient(client)
			}
			t.clients[client] = true
			go t.GoRecv(client, onRecv)
		case client := <-t.offClients:
			if ok := t.clients[client]; ok {
				libol.Debug("TcpServer.delClient %s", client.Addr)
				if onClose != nil {
					onClose(client)
				}
				client.Close()
				delete(t.clients, client)
			}
		}
	}
}

func (t *TcpServer) GoRecv(client *libol.TcpClient, onRecv func(*libol.TcpClient, []byte) error) {
	libol.Debug("TcpServer.GoRecv: %s", client.Addr)
	for {
		data := make([]byte, 4096)
		length, err := client.RecvMsg(data)
		if err != nil {
			t.offClients <- client
			break
		}

		if length > 0 {
			libol.Debug("TcpServer.GoRecv: length: %d ", length)
			libol.Debug("TcpServer.GoRecv: data  : % x", data[:length])
			onRecv(client, data[:length])
		}
	}
}
