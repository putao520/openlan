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
    verbose bool
    newtime int64

    //Public variable
    TxOkay uint64
    RxOkay uint64
    TxError uint64
    Droped uint64
    MaxSize int
    MinSize int
}

func NewUdpSocket(addr string, verbose bool) (this *UdpSocket) {
    this = &UdpSocket {
        addr: addr,
        conn: nil,
        MaxSize: 1514,
        MinSize: int(HSIZE),
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
//
// [MAGIC(2)][Length(2)]
// [UUID(16)]
func (this *UdpSocket) SendMsg(addr *net.UDPAddr, uuid string, data []byte) error {
    if !this.IsOk() {
        return errors.New("Connection isn't okay!")
    }

    buffer := make([]byte, 20+len(data))

    copy(buffer[:2], MAGIC)
    binary.BigEndian.PutUint16(buffer[2:4], uint16(len(data)))
    if len(uuid) == 16 {
        copy(buffer[4:20], []byte(uuid))
    } else {
        copy(buffer[4:20], CTLUUID)
    }
    copy(buffer[20:], data)

    if this.verbose {
        log.Printf("Debug| UdpSocket.SendMsg to %s,%s %d: % x...", 
                    addr, uuid, len(buffer), buffer[:16])
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

func (this *UdpSocket) RecvMsg() (*net.UDPAddr, string, []byte, error) {
    if !this.IsOk() {
        return nil, "", nil, errors.New("Connection isn't okay!")
    }

    var err error
    var addr *net.UDPAddr

    n := 0
    data := make([]byte, this.MaxSize)
    for {
        n, addr, err = this.conn.ReadFromUDP(data)    
        if err != nil {
            return nil, "", nil, err
        }

        data = data[:n]
        break
    }

    head := data[:20]
    if !bytes.Equal(head[0:2], MAGIC) {
        return nil, "", nil, errors.New("Isn't right magic header!")
    }
    uuid := head[4:20]
    size := binary.BigEndian.Uint16(head[2:4])
    if this.verbose {
        log.Printf("Debug| UdpSocket.RecvMsg from %s,%s %d. % x", addr, uuid, n, data[:32])
    }

    if int(size) > this.MaxSize || int(size) < this.MinSize {
        return nil, "", nil, errors.New(fmt.Sprintf("Isn't right data size(%d)!", size))
    }

    this.RxOkay++

    return addr, string(uuid), data[HSIZE:], nil
}

func (this *UdpSocket) IsOk() bool {
    return this.conn != nil
}

func (this *UdpSocket) GetAddr() string {
    return this.addr
}

func (this *UdpSocket) SendReq(addr *net.UDPAddr, uuid string, mesg *Message) error {
    data := mesg.EncodeReq()

    if this.verbose {
    	log.Printf("Debug| UdpSocket.SendReq %d %s\n", len(data), data[6:])
    }

	if err := this.SendMsg(addr, uuid, data); err != nil {
		return err
    }
    return nil
}

func (this *UdpSocket) SendResp(addr *net.UDPAddr, uuid string, mesg *Message) error {
    data := mesg.EncodeResp()

    if this.verbose {
    	log.Printf("Debug| UdpSocket.SendResp %d %s\n", len(data), data[6:])
    }

	if err := this.SendMsg(addr, uuid, data); err != nil {
		return err
    }
    return nil
}

func (this *UdpSocket) UpTime() int64 {
    return time.Now().Unix() - this.newtime
}
