package openlan

import (
    "net"
    "fmt"
    "errors"
    "log"
)

type TcpServer struct {
	addr string
	port string
	conn net.Conn
	maxClient int
	clients map[*Client]bool
	onClients chan *Client
	offClients chan *Client
}

func NewTcpServer(addr string, port string) (this *TcpServer) {
	this = &TcpServer {
		addr: addr,
		port: port,
		conn: nil,
		maxClient: 1024,
		clients: make(map[*Client]bool),
		onClients: make(*Client, 4),
		offClients: make(*Client, 8),
	}

	this.Listen()

	return 
}

func (this *TcpServer) (Listen) (err error) {
	log.Printf("TcpServer.Start %s:%d\n", this.addr, this.port)

	this.conn, err = net.Listen("tcp", fmt.Sprintf("%s:%d", this.addr, this.port))
	if err != nil {
		log.Print("TcpServer.Listen: %s", err)
		this.conn = nil
        return
    }
}

func (this *TcpServer) GoAccept() error {
	for {
		conn, err := this.conn.Accept()
		if err != nil {
			log.Print("TcpServer.GoAccept: %s", err)
		}

		this.onClients <- newClient(conn)
	}
}

func (this *TcpServer) GoLoop(client *Client) {
	for {
		select {
		case client := <- this.onClients:
			log.Print("TcpServer.addClient %s", client.conn)
			client.Open()
			this.clients[client] = true

			go this.GoRecv(client)
		case client := <- this.offClients:
			if ok := manager.clients[client]; ok {
				log.Print("TcpServer.delClient %s", client.conn)
				client.Close()
				delete(this.clients, client)
			}
		}
	}

}

func (this *TcpServer) GoRecv(client *Client) {
    for {
        data := make([]byte, 4096)
        length, err := client.RecvMsg(data)
        if err != nil {
            manager.offClients <- client
            break
		}
		
        if length > 0 {
            fmt.Println("RECEIVED: " + string(data))
            //TODO send to TAP device.
        }
    }
}