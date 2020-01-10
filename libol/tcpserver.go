package libol

import (
	"crypto/tls"
	"net"
)

type TcpServerListener struct {
	OnClient func(client *TcpClient) error
	OnClose func(client *TcpClient) error
	ReadAt func(client *TcpClient, p []byte) error
}

type TcpServerSts struct {
	RxCount  int64
	TxCount  int64
	DrpCount int64
	AcpCount int64
	ClsCount int64
}

type TcpServer struct {
	Addr     string
	TlsConf  *tls.Config
	Sts      TcpServerSts

	listener   net.Listener
	maxClient  int
	clients    map[*TcpClient]bool
	onClients  chan *TcpClient
	offClients chan *TcpClient
}

func NewTcpServer(listen string, config *tls.Config) (t *TcpServer) {
	t = &TcpServer{
		Addr:       listen,
		TlsConf:    config,
		Sts:        TcpServerSts{},
		maxClient:  1024,
		clients:    make(map[*TcpClient]bool, 1024),
		onClients:  make(chan *TcpClient, 4),
		offClients: make(chan *TcpClient, 8),
	}

	if err := t.Listen(); err != nil {
		Debug("NewTcpServer %s", err)
	}

	return
}

func (t *TcpServer) Listen() (err error) {
	Info("TcpServer.Start %s,%p", t.Addr, t.TlsConf)

	if t.TlsConf != nil {
		t.listener, err = tls.Listen("tcp", t.Addr, t.TlsConf)
		if err != nil {
			Info("TcpServer.Listen: %s", err)
			return
		}
	} else {
		t.listener, err = net.Listen("tcp", t.Addr)
		if err != nil {
			Info("TcpServer.Listen: %s", err)
			t.listener = nil
			return
		}
	}

	return nil
}

func (t *TcpServer) Close() {
	if t.listener != nil {
		t.listener.Close()
		Info("TcpServer.Close: %s", t.Addr)
		t.listener = nil
	}
}

func (t *TcpServer) Accept() {
	Debug("TcpServer.Accept")
	if t.listener == nil {
		Error("TcpServer.Accept: invalid listener")
	}

	defer t.Close()
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			Error("TcpServer.Accept: %s", err)
			return
		}

		t.Sts.AcpCount++
		t.onClients <- NewTcpClientFromConn(conn)
	}
}

func (t *TcpServer) Loop(call TcpServerListener) {
	Debug("TcpServer.Loop")
	defer t.Close()
	for {
		select {
		case client := <-t.onClients:
			Debug("TcpServer.addClient %s", client.Addr)
			t.clients[client] = true
			if call.OnClient != nil {
				call.OnClient(client)
				if call.ReadAt != nil {
					go t.Read(client, call.ReadAt)
				}
			}
		case client := <-t.offClients:
			if ok := t.clients[client]; ok {
				Debug("TcpServer.delClient %s", client.Addr)
				t.Sts.ClsCount++
				if call.OnClose != nil {
					call.OnClose(client)
				}
				client.Close()
				delete(t.clients, client)
			}
		}
	}
}

func (t *TcpServer) Read(client *TcpClient, ReadAt func(client *TcpClient, p []byte) error) {
	Debug("TcpServer.Read: %s", client.Addr)
	for {
		data := make([]byte, 4096)
		length, err := client.ReadMsg(data)
		if err != nil {
			Error("TcpServer.Read: %s", err)
			t.offClients <- client
			break
		}

		if length > 0 {
			t.Sts.RxCount++
			Debug("TcpServer.Read: length: %d ", length)
			Debug("TcpServer.Read: data  : % x", data[:length])
			ReadAt(client, data[:length])
		}
	}
}
