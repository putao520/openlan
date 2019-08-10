package openlan

import (
    "net"
    "fmt"
    "errors"
    "encoding/binary"
    "bytes"
    "log"
)

var (
    MAGIC  = []byte{0xff,0xff}
    HSIZE  = uint16(0x04)
)

type TcpClient struct {
    addr string
    port uint16
    conn net.Conn
    maxsize int
    minsize int
    verbose int
}

func NewTcpClient(addr string, port uint16, verbose int) (this *TcpClient, err error){
    this = &TcpClient {
        addr: addr,
        port: port,
        conn: nil,
        maxsize: 1514,
        minsize: 15,
        verbose: verbose,
    }

    err = this.Connect()
    return 
}

func (this *TcpClient) Connect() (err error) {
    if this.conn != nil {
        return nil
    }

    if this.IsVerbose() {
        log.Printf("TcpClient.Connect to %s:%d\n", this.addr, this.port)
    }

    this.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", this.addr, this.port))
    if err != nil {
        this.conn = nil
        return
    }
    return nil
}

func (this *TcpClient) Close() {
    if this.conn != nil {
        this.conn.Close()
    }
}

func (this *TcpClient) recvn(buffer []byte) error {
    offset := 0
    left := len(buffer)
    for left > 0 {
        tmp := make([]byte, left)
        n, err := this.conn.Read(tmp)
        if err != nil {
            return err
        }

        copy(buffer[offset:], tmp)

        offset += n
        left -= n 
    }
    
    if this.IsVerbose() {
        log.Printf("TcpClient.recvn %d\n", len(buffer))
        log.Printf("TcpClient.recvn Data: % x\n", buffer)
    }

    return nil
}

func (this *TcpClient) sendn(buffer []byte) error {
    offset := 0
    size := len(buffer)
    left := size - offset
    if this.IsVerbose() {
        log.Printf("TcpClient.sendn %d\n", size)
        log.Printf("TcpClient.sendn Data: % x\n", buffer)
    }

    for left > 0 {
        tmp := buffer[offset:]
        if this.IsVerbose() {
            log.Printf("TcpClient.sendn tmp %d\n", len(tmp))
        }
        n, err := this.conn.Write(tmp)
        if err != nil {
            return err 
        }
        offset += n
        left = size - offset
    }
    return nil
}

func (this *TcpClient) SendMsg(data []byte) error {
    if err := this.Connect(); err != nil {
        return err
    }

    buffer := make([]byte, int(HSIZE)+len(data))
    copy(buffer[0:2], MAGIC)
    binary.BigEndian.PutUint16(buffer[2:4], uint16(len(data)))
    copy(buffer[HSIZE:], data)

    return this.sendn(buffer)
}

func (this *TcpClient) RecvMsg(data []byte) (int, error) {
    h := make([]byte, HSIZE)
    if err := this.recvn(h); err != nil {
        return -1, err
    }

    if !bytes.Equal(h[0:2], MAGIC) {
        return -1, errors.New("Isn't right magic header!")
    }

    size := binary.BigEndian.Uint16(h[2:4])
    if int(size) > this.maxsize || int(size) < this.minsize {
        return -1, errors.New(fmt.Sprintf("Isn't right data size(%d)!", size))
    }

    d := make([]byte, size)
    if err := this.recvn(d); err != nil {
        return -1, err
    }

    copy(data, d)

    return int(size), nil
}

func (this *TcpClient) IsVerbose() bool {
    return this.verbose != 0
}

func (this *TcpClient) GetMaxSize() int {
    return this.maxsize
}

func (this *TcpClient) SetMaxSize(value int) {
    this.maxsize = value
}

func (this *TcpClient) GetMinSize() int {
    return this.minsize
}