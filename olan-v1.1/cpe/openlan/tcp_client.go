package openlan

import (
	"net"
	"fmt"
	"errors"
	"encoding/binary"
	"bytes"
)

var (
	MAGIC  = []byte{0xff,0xff}
	HSIZE  = uint16(0x04)
)

type TcpClient struct {
	addr string
	port uint16
	conn net.Conn
	maxsize uint32
	minsize uint32
}

func NewTcpClient(addr string, port uint16) (client *TcpClient, err error){
	client = &TcpClient{
		addr: addr,
		port: port,
		conn: nil,
		maxsize: 1514,
		minsize: 15,
	}

	err = client.Connect()
	return 
}

func (this *TcpClient) Connect() (err error) {
	if this.conn != nil {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", this.addr, this.port)
	this.conn, err = net.Dial("tcp", addr)
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
	return nil
}

func (this *TcpClient) sendn(buffer []byte) error {
	offset := 0
	size := len(buffer)
	// fmt.Printf("sendn %d\n", size)
	// fmt.Printf("Data: % x\n", buffer)
	for size - offset >= 1 {
		tmp := buffer[offset:]
		// fmt.Printf("sendn tmp %d\n", len(tmp))
		n, err := this.conn.Write(tmp)
		if err != nil {
			return err 
		}
		offset += n
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
	if uint32(size) > this.maxsize || uint32(size) < this.minsize {
		return -1, errors.New(fmt.Sprintf("Isn't right data size(%d)!", size))
	}

	d := make([]byte, size)
	if err := this.recvn(d); err != nil {
		return -1, err
	}

	copy(data, d)

	return int(size), nil
}