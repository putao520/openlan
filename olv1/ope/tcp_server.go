package olv1ope

import (
    "net"
	"log"
	
	"github.com/danieldin95/openlan-go/olv1/olv1"
)

type TcpServer struct {
	addr string
	listener *net.TCPListener
	maxClient int
	clients map[*olv1.TcpClient]bool
	onClients chan *olv1.TcpClient
	offClients chan *olv1.TcpClient
	verbose int
}

func NewTcpServer(c *Config) (this *TcpServer) {
	this = &TcpServer {
		addr: c.TcpListen,
		listener: nil,
		maxClient: 1024,
		clients: make(map[*olv1.TcpClient]bool, 1024),
		onClients: make(chan *olv1.TcpClient, 4),
		offClients: make(chan *olv1.TcpClient, 8),
		verbose: c.Verbose,
	}

	if err := this.Listen(); err != nil {
		log.Printf("NewTcpServer %s\n", err)
	}

	return 
}

func (this *TcpServer) Listen() error {
	log.Printf("TcpServer.Start %s\n", this.addr)

	laddr, err := net.ResolveTCPAddr("tcp", this.addr)
    if err != nil {
        return err
	}
	
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Printf("TcpServer.Listen: %s", err)
		this.listener = nil
        return err
	}
	this.listener = listener
	return nil
}

func (this *TcpServer) Close() {
	if this.listener != nil {
		this.listener.Close()
		log.Printf("TcpServer.Close: %s", this.addr)
		this.listener = nil
	}
}

func (this *TcpServer) GoAccept() {
	log.Printf("TcpServer.GoAccept")
	if (this.listener == nil) {
		log.Printf("Error|TcpServer.GoAccept: invalid listener")
	}

	defer this.Close()
	for {
		conn, err := this.listener.AcceptTCP()
		if err != nil {
			log.Printf("Error|TcpServer.GoAccept: %s", err)
			return
		}

		this.onClients <- olv1.NewTcpClientFromConn(conn, this.verbose)
	}

	return
}

func (this *TcpServer) GoLoop(onClient func (*olv1.TcpClient) error, 
							  onRecv func (*olv1.TcpClient, []byte) error,
							  onClose func (*olv1.TcpClient) error) {
	log.Printf("TcpServer.GoLoop")
	defer this.Close()
	for {
		select {
		case client := <- this.onClients:
			log.Printf("TcpServer.addClient %s", client.GetAddr())
			if onClient != nil {
				onClient(client)
			}
			this.clients[client] = true
			go this.GoRecv(client, onRecv)
		case client := <- this.offClients:
			if ok := this.clients[client]; ok {
				log.Printf("TcpServer.delClient %s", client.GetAddr())
				if onClose != nil {
					onClose(client)
				}
				client.Close()
				delete(this.clients, client)
			}
		}
	}
}

func (this *TcpServer) GoRecv(client *olv1.TcpClient, onRecv func (*olv1.TcpClient, []byte) error) {
	log.Printf("TcpServer.GoRecv: %s", client.GetAddr())	
    for {
        data := make([]byte, 4096)
        length, err := client.RecvMsg(data)
        if err != nil {
            this.offClients <- client
            break
		}
		
        if length > 0 {
			if this.IsVerbose() {
				log.Printf("TcpServer.GoRecv: length: %d ", length)
				log.Printf("TcpServer.GoRecv: data  : % x", data[:length])
			}
			onRecv(client, data[:length])
        }
    }
}

func (this *TcpServer) GetAddr() string {
	return this.addr
}

func (this *TcpServer) IsVerbose() bool {
	return this.verbose != 0
}