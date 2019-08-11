package openlan

import (
    "net"
    "fmt"
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

func NewTcpServer(addr string) (this *TcpServer) {
	this = &TcpServer {
		addr: addr,
		listener: nil,
		maxClient: 1024,
		clients: make(map[*TcpClient]bool),
		onClients: make(chan *TcpClient, 4),
		offClients: make(chan *TcpClient, 8),
		verbose: 1,
	}

	this.Listen()

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
		log.Print("TcpServer.Listen: %s", err)
		this.listener = nil
        return err
	}
	this.listener = listener
	return nil
}

func (this *TcpServer) GoAccept() error {
	for {
		conn, err := this.listener.AcceptTCP()
		if err != nil {
			log.Print("TcpServer.GoAccept: %s", err)
		}

		this.onClients <- NewTcpClientFromConn(conn, this.verbose)
	}
}

func (this *TcpServer) GoLoop(client *TcpClient) {
	for {
		select {
		case client := <- this.onClients:
			log.Print("TcpServer.addClient %s", client.conn)
			// client.Open()
			this.clients[client] = true

			go this.GoRecv(client)
		case client := <- this.offClients:
			if ok := this.clients[client]; ok {
				log.Print("TcpServer.delClient %s", client.conn)
				client.Close()
				delete(this.clients, client)
			}
		}
	}

}

func (this *TcpServer) GoRecv(client *TcpClient) {
    for {
        data := make([]byte, 4096)
        length, err := client.RecvMsg(data)
        if err != nil {
            this.offClients <- client
            break
		}
		
        if length > 0 {
            fmt.Println("RECEIVED: " + string(data))
            //TODO send to TAP device.
        }
    }
}