package olv1

import (
    "net"
    "log"
)

type TcpServer struct {
	addr string
	listener *net.TCPListener
	maxClient int
	clients map[*TcpClient]bool
	onClients chan *TcpClient
	offClients chan *TcpClient
	verbose int
}

func NewTcpServer(addr string, verbose int) (this *TcpServer) {
	this = &TcpServer {
		addr: addr,
		listener: nil,
		maxClient: 1024,
		clients: make(map[*TcpClient]bool),
		onClients: make(chan *TcpClient, 4),
		offClients: make(chan *TcpClient, 8),
		verbose: verbose,
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
	}
}

func (this *TcpServer) GoAccept() error {
	log.Printf("TcpServer.GoAccept")
	if (this.listener == nil) {
		log.Printf("Error|TcpServer.GoAccept: invalid listener")
		return nil
	}

	defer this.Close()
	for {
		conn, err := this.listener.AcceptTCP()
		if err != nil {
			log.Printf("Error|TcpServer.GoAccept: %s", err)
		}

		this.onClients <- NewTcpClientFromConn(conn, this.verbose)
	}
}

func (this *TcpServer) GoLoop(onClient func (*TcpClient) error, 
							  onRecv func (*TcpClient, []byte) error,
							  onClose func (*TcpClient) error) {
	log.Printf("TcpServer.GoLoop")
	defer this.Close()
	for {
		select {
		case client := <- this.onClients:
			log.Printf("TcpServer.addClient %s", client.conn)
			if onClient != nil {
				onClient(client)
			}
			this.clients[client] = true
			go this.GoRecv(client, onRecv)
		case client := <- this.offClients:
			if ok := this.clients[client]; ok {
				log.Printf("TcpServer.delClient %s", client.conn)
				if onClose != nil {
					onClose(client)
				}
				client.Close()
				delete(this.clients, client)
			}
		}
	}
}

func (this *TcpServer) GoRecv(client *TcpClient, onRecv func (*TcpClient, []byte) error) {
	log.Printf("TcpServer.GoRecv: %s", client)	
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