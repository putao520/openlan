package vswitch

import (
    "net"
    "log"
    
    "github.com/danieldin95/openlan-go/olv1/openlanv1"
)

type TcpServer struct {
    addr string
    listener *net.TCPListener
    maxClient int
    clients map[*openlanv1.TcpClient]bool
    onClients chan *openlanv1.TcpClient
    offClients chan *openlanv1.TcpClient
    verbose int
}

func NewTcpServer(c *Config) (this *TcpServer) {
    this = &TcpServer {
        addr: c.TcpListen,
        listener: nil,
        maxClient: 1024,
        clients: make(map[*openlanv1.TcpClient]bool, 1024),
        onClients: make(chan *openlanv1.TcpClient, 4),
        offClients: make(chan *openlanv1.TcpClient, 8),
        verbose: c.Verbose,
    }

    if err := this.Listen(); err != nil {
        log.Printf("Debug| NewTcpServer %s\n", err)
    }

    return 
}

func (this *TcpServer) Listen() error {
    log.Printf("Debug| TcpServer.Start %s\n", this.addr)

    laddr, err := net.ResolveTCPAddr("tcp", this.addr)
    if err != nil {
        return err
    }
    
    listener, err := net.ListenTCP("tcp", laddr)
    if err != nil {
        log.Printf("Info| TcpServer.Listen: %s", err)
        this.listener = nil
        return err
    }
    this.listener = listener
    return nil
}

func (this *TcpServer) Close() {
    if this.listener != nil {
        this.listener.Close()
        log.Printf("Info| TcpServer.Close: %s", this.addr)
        this.listener = nil
    }
}

func (this *TcpServer) GoAccept() {
    log.Printf("Debug| TcpServer.GoAccept")
    if (this.listener == nil) {
        log.Printf("Error| TcpServer.GoAccept: invalid listener")
    }

    defer this.Close()
    for {
        conn, err := this.listener.AcceptTCP()
        if err != nil {
            log.Printf("Error| TcpServer.GoAccept: %s", err)
            return
        }

        this.onClients <- openlanv1.NewTcpClientFromConn(conn, this.verbose)
    }

    return
}

func (this *TcpServer) GoLoop(onClient func (*openlanv1.TcpClient) error, 
                              onRecv func (*openlanv1.TcpClient, []byte) error,
                              onClose func (*openlanv1.TcpClient) error) {
    log.Printf("Debug| TcpServer.GoLoop")
    defer this.Close()
    for {
        select {
        case client := <- this.onClients:
            log.Printf("Debug| TcpServer.addClient %s", client.GetAddr())
            if onClient != nil {
                onClient(client)
            }
            this.clients[client] = true
            go this.GoRecv(client, onRecv)
        case client := <- this.offClients:
            if ok := this.clients[client]; ok {
                log.Printf("Debug| TcpServer.delClient %s", client.GetAddr())
                if onClose != nil {
                    onClose(client)
                }
                client.Close()
                delete(this.clients, client)
            }
        }
    }
}

func (this *TcpServer) GoRecv(client *openlanv1.TcpClient, onRecv func (*openlanv1.TcpClient, []byte) error) {
    log.Printf("Debug| TcpServer.GoRecv: %s", client.GetAddr())    
    for {
        data := make([]byte, 4096)
        length, err := client.RecvMsg(data)
        if err != nil {
            this.offClients <- client
            break
        }
        
        if length > 0 {
            if this.IsVerbose() {
                log.Printf("Debug| TcpServer.GoRecv: length: %d ", length)
                log.Printf("Debug| TcpServer.GoRecv: data  : % x", data[:length])
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