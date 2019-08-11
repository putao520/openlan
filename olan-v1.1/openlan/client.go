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

type Client struct {
    conn net.Conn
	verbose int
	maxSize int
	minSzie int
}

func NewClient(conn net.Conn) (this *Client) {
	this = &Client {
		conn: conn,
		verbose: 0,
		maxSize: 1514,
		minSize: 15,
	}
	
	return 
}

func (this *Client) Open() {
}

func (this *Client) Close() {
	this.Close()
}

func (this *Client) recvn(buffer []byte) error {
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
        log.Printf("Client.recvn %d\n", len(buffer))
        log.Printf("Client.recvn Data: % x\n", buffer)
    }

    return nil
}

func (this *Client) sendn(buffer []byte) error {
    offset := 0
    size := len(buffer)
    left := size - offset
    if this.IsVerbose() {
        log.Printf("Client.sendn %d\n", size)
        log.Printf("Client.sendn Data: % x\n", buffer)
    }

    for left > 0 {
        tmp := buffer[offset:]
        if this.IsVerbose() {
            log.Printf("Client.sendn tmp %d\n", len(tmp))
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

func (this *Client) SendMsg(data []byte) error {
    if err := this.Connect(); err != nil {
        return err
    }

    buffer := make([]byte, int(HSIZE)+len(data))
    copy(buffer[0:2], MAGIC)
    binary.BigEndian.PutUint16(buffer[2:4], uint16(len(data)))
    copy(buffer[HSIZE:], data)

    return this.sendn(buffer)
}

func (this *Client) RecvMsg(data []byte) (int, error) {
    if !this.IsOk() {
        return -1, errors.New("Connection isn't okay!")
    }

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

func (this *Client) IsVerbose() bool {
    return this.verbose != 0
}