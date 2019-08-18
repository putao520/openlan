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
    verbose int
    newtime int64
    //Public variable
    TxOkay uint64
    RxOkay uint64
    TxError uint64
    Droped uint64
}

func UdpSocket(addr string, verbose int) (this *UdpSocket) {
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

func (this *UdpSocket) Close() {
    if this.conn != nil {
        log.Printf("UdpSocket.Close %s\n", this.addr)
        this.conn.Close()
        this.conn = nil
    }
}

func (this *UdpSocket) Recvn(buffer []byte) (addr *net.UDPAddr, err error) {
    offset := 0
    left := len(buffer)
    for left > 0 {
        n := 0
        tmp := make([]byte, left)
        n, addr, err = this.conn.ReadFromUDP(tmp)
        if err != nil {
            return 
        }

        copy(buffer[offset:], tmp)

        offset += n
        left -= n 
    }
    
    if this.IsVerbose() {
        log.Printf("UdpSocket.recvn %d\n", len(buffer))
        log.Printf("UdpSocket.recvn Data: % x\n", buffer)
    }

    return 
}

func (this *UdpSocket) Sendn(addr *net.UDPAddr, buffer []byte) error {
    offset := 0
    size := len(buffer)
    left := size - offset
    if this.IsVerbose() {
        log.Printf("UdpSocket.sendn %d\n", size)
        log.Printf("UdpSocket.sendn Data: % x\n", buffer)
    }

    for left > 0 {
        tmp := buffer[offset:]
        if this.IsVerbose() {
            log.Printf("UdpSocket.sendn tmp %d\n", len(tmp))
        }
        n, err := this.conn.WriteToUDP(tmp, addr)
        if err != nil {
            return err 
        }
        offset += n
        left = size - offset
    }
    return nil
}

func (this *UdpSocket) SendMsg(addr *net.UDPAddr, data []byte) error {
    buffer := make([]byte, int(HSIZE)+len(data))
    copy(buffer[0:2], MAGIC)
    binary.BigEndian.PutUint16(buffer[2:4], uint16(len(data)))
    copy(buffer[HSIZE:], data)

    if err := this.Sendn(buffer, addr); err != nil {
        this.TxError++
        return err
    }
    
    this.TxOkay++

    return nil
}

func (this *UdpSocket) RecvMsg() (net.UDPAddr, []byte, error) {
    if !this.IsOk() {
        return -1, errors.New("Connection isn't okay!")
    }

    head := make([]byte, HSIZE)
    addr, err := this.Recvn(head)
    if err != nil {
        return nil, nil, err
    }

    if !bytes.Equal(h[0:2], MAGIC) {
        return nil, nil, errors.New("Isn't right magic header!")
    }

    size := binary.BigEndian.Uint16(h[2:4])
    if int(size) > this.maxsize || int(size) < this.minsize {
        return nil, nil, errors.New(fmt.Sprintf("Isn't right data size(%d)!", size))
    }

    data := make([]byte, size)
    addr, err := this.Recvn(d)
    if err != nil {
        return nil, nil, err
    }

    this.RxOkay++

    return addr, data, nil
}

func (this *UdpSocket) IsVerbose() bool {
    return this.verbose != 0
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

    if this.IsVerbose() {
    	log.Printf("Debug| UdpSocket.SendReq %d %s\n", len(data), data[6:])
    }

	if err := this.SendMsg(addr, data); err != nil {
		return err
    }
    return nil
}

func (this *UdpSocket) SendResp(addr *net.UDPAddr, action string, body string) error {
    data := EncInstResp(action, body)

    if this.IsVerbose() {
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
