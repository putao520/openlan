package olv1

import (
	"log"
	"time"
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
	defer this.client.Close()
	for {
		if !this.client.IsOk() {
			time.Sleep(2 * time.Second) // sleep 2s to release cpu.
			continue
		}

		data := make([]byte, this.maxSize)
        n, err := this.client.RecvMsg(data)
        if err != nil {
			log.Printf("Error|TcpWroker.GoRev: %s", err)
			this.client.Close()
			continue
		}
		if this.IsVerbose() {
			log.Printf("TcpWroker.GoRev: % x\n", data[:n])
		}
	
		if n > 0 {
			dorecv(data[:n])
		}
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
	defer this.client.Close()
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