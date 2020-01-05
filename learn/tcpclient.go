package main

import (
	"encoding/binary"
	"fmt"
	"github.com/songgao/water"
	"net"
)

func ReadFull(conn net.Conn, buffer []byte) error {
	offset := 0
	left := len(buffer)

	for left > 0 {
		tmp := make([]byte, left)
		n, err := conn.Read(tmp)
		if err != nil {
			return err
		}
		copy(buffer[offset:], tmp)
		offset += n
		left -= n
	}
	return nil
}

func WriteFull(conn net.Conn, buffer []byte) error {
	offset := 0
	size := len(buffer)
	left := size - offset

	for left > 0 {
		tmp := buffer[offset:]
		n, err := conn.Write(tmp)
		if err != nil {
			return err
		}
		offset += n
		left = size - offset
	}
	return nil
}

func main() {
	sip := net.ParseIP("192.168.209.141")
	srcAddr := &net.TCPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.TCPAddr{IP: sip, Port: 9981}

	conn, err := net.DialTCP("tcp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	device, err := water.New(water.Config{DeviceType: water.TAP})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Local: <%s> \n", device.Name())

	go func() {
		frameData := make([]byte, 1600+4)

		for {
			n, err := device.Read(frameData[4:])
			if err != nil {
				break
			}
			if n == 0 || conn == nil {
				continue
			}

			binary.BigEndian.PutUint16(frameData[2:4], uint16(n))
			//fmt.Printf("<%s> %d\n", device.Name(), n)
			//fmt.Printf("<%s> % x\n", device.Name(), frameData[:20])
			err = WriteFull(conn, frameData[:n+4])
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	for {
		data := make([]byte, 1600+4)

		err := ReadFull(conn, data[:4])
		if err != nil {
			fmt.Printf("error during read: %s", err)
		}

		size := binary.BigEndian.Uint16(data[2:4])
		if size == 0 || size > 1600 {
			continue
		}

		err = ReadFull(conn, data[4:size+4])
		if err != nil {
			fmt.Printf("error during read: %s", err)
		}

		_, err = device.Write(data[4:size+4])
		if err != nil {
			fmt.Println(err)
		}
	}

	conn.Close()
	device.Close()
}