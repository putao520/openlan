package openlan

import (
	"log"
)

type TcpWroker struct {
	client *TcpClient
	readchan chan []byte
	writechan chan []byte
	verbose int
	maxSize int
}

func NewTcpWoker(client *TcpClient, maxSize int, verbose int) (this *TcpWroker) {
	this = &TcpWroker {
		client: client,
		writechan: make(chan []byte, 1024*10),
		verbose: verbose,
		maxSize: maxSize,
	}
	this.client.SetMaxSize(this.maxSize)

	return
}

func (this *TcpWroker) GoRecv(dorecv func([]byte)(error)) {
	for {
		data := make([]byte, this.maxSize)
        n, err := this.client.RecvMsg(data)
        if err != nil || n <= 0 {
			log.Printf("Error|TcpWroker.GoRev: %s", err)
			continue
		}
		if this.IsVerbose() {
			log.Printf("TcpWroker.GoRev: % x\n", data[:n])
		}
	
		dorecv(data[:n])
	}
}

func (this *TcpWroker) DoSend(data []byte) error {
	if this.IsVerbose() {
		log.Printf("TcpWroker.DoSend: % x\n", data)
	}

	this.writechan <- data
	return nil
}

func (this *TcpWroker) GoLoop() error {
	for {
		select {
		case wdata := <- this.writechan:
			if err := this.client.SendMsg(wdata); err != nil {
				log.Printf("Error|TcpWroker.GoLoop: %s", err)
			}
		}
	}
}

func (this *TcpWroker) IsVerbose() bool {
    return this.verbose != 0
}