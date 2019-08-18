package openlanv2

import (
    "net"
    "fmt"
    "errors"
    "encoding/binary"
    "bytes"
    "log"
    "time"
)

type UdpSocket struct {
    addr string
    conn *net.UDPConn
    maxsize int
    minsize int
    verbose bool
    newtime int64
    //Public variable
    TxOkay uint64
    RxOkay uint64
    TxError uint64
    Droped uint64
}

func NewUdpSocket(addr string, verbose bool) (this *UdpSocket) {
    this = &UdpSocket {
        addr: addr,
        conn: nil,
        maxsize: 1514,
        minsize: 15,
        verbose: verbose,
        TxOkay: 0,
        RxOkay: 0,
        TxError: 0,
        Droped: 0,
        newtime: time.Now().Unix(),
    }
    return 
}

func (this *UdpSocket) Listen() error {
    if this.conn != nil {
        return nil
    }

    log.Printf("UdpSocket.Listen %s\n", this.addr)
    laddr, err := net.ResolveUDPAddr("udp", this.addr)
    if err != nil {
        return err
    }

    conn, err := net.ListenUDP("udp", laddr)
    if err != nil {
        this.conn = nil
        return err
    }
    this.conn = conn

    return nil
}

func (this *UdpSocket) Close() (err error) {
    if this.conn != nil {
        log.Printf("UdpSocket.Close %s\n", this.addr)
        err = this.conn.Close()
        this.conn = nil
    }
    return 
}

func (this *UdpSocket) SendMsg(addr *net.UDPAddr, data []byte) error {
    if !this.IsOk() {
        return errors.New("Connection isn't okay!")
    }

    buffer := make([]byte, int(HSIZE)+len(data))
    copy(buffer[0:2], MAGIC)
    binary.BigEndian.PutUint16(buffer[2:4], uint16(len(data)))
    copy(buffer[HSIZE:], data)

    if this.verbose {
        log.Printf("Debug| UdpSocket.SendMsg %d. % x", len(buffer), buffer)
    }

    n, err := this.conn.WriteToUDP(buffer, addr)
    if err != nil {
        this.TxError++
        return err
    }

    if this.verbose {
        log.Printf("Debug| UdpSocket.SendMsg %d.", n)
    }

    this.TxOkay++

    return nil
}

func (this *UdpSocket) RecvMsg() (*net.UDPAddr, []byte, error) {
    if !this.IsOk() {
        return nil, nil, errors.New("Connection isn't okay!")
    }

    var err error
    var addr *net.UDPAddr

    n := 0
    data := make([]byte, this.maxsize)
    for {
        n, addr, err = this.conn.ReadFromUDP(data)    
        if err != nil {
            return nil, nil, err
        }

        data = data[:n]
        break
    }

    if this.verbose {
        log.Printf("Debug| UdpSocket.RecvMsg %d. % x", n, data)
    }

    head := data[:HSIZE]
    if !bytes.Equal(head[0:2], MAGIC) {
        return nil, nil, errors.New("Isn't right magic header!")
    }

    size := binary.BigEndian.Uint16(head[2:4])
    if int(size) > this.maxsize || int(size) < this.minsize {
        return nil, nil, errors.New(fmt.Sprintf("Isn't right data size(%d)!", size))
    }

    this.RxOkay++

    return addr, data[HSIZE:], nil
}

func (this *UdpSocket) GetMaxSize() int {
    return this.maxsize
}

func (this *UdpSocket) SetMaxSize(value int) {
    this.maxsize = value
}

func (this *UdpSocket) GetMinSize() int {
    return this.minsize
}

func (this *UdpSocket) IsOk() bool {
    return this.conn != nil
}

func (this *UdpSocket) GetAddr() string {
    return this.addr
}

func (this *UdpSocket) SendReq(addr *net.UDPAddr, action string, body string) error {
    data := EncInstReq(action, body)

    if this.verbose {
    	log.Printf("Debug| UdpSocket.SendReq %d %s\n", len(data), data[6:])
    }

	if err := this.SendMsg(addr, data); err != nil {
		return err
    }
    return nil
}

func (this *UdpSocket) SendResp(addr *net.UDPAddr, action string, body string) error {
    data := EncInstResp(action, body)

    if this.verbose {
    	log.Printf("Debug| UdpSocket.SendResp %d %s\n", len(data), data[6:])
    }

	if err := this.SendMsg(addr, data); err != nil {
		return err
    }
    return nil
}

func (this *UdpSocket) UpTime() int64 {
    return time.Now().Unix() - this.newtime
}
