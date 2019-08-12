package openlanv1

import (
	"log"

	"github.com/songgao/water"
)

type TapWroker struct {
	ifce *water.Interface
	writechan chan []byte
	ifmtu int
	verbose int
}

func NewTapWoker(ifce *water.Interface, ifmtu int, verbose int) (this *TapWroker) {
	this = &TapWroker {
		ifce: ifce,
		writechan: make(chan []byte, 1024*10),
		ifmtu: ifmtu, //1514
		verbose: verbose,
	}
	return
}

func (this *TapWroker) GoRecv(dorecv func([]byte)(error)) {
	defer this.ifce.Close()
	for {
		data := make([]byte, this.ifmtu)
        n, err := this.ifce.Read(data)
        if err != nil {
			log.Printf("Error|TapWroker.GoRev: %s", err)
			continue
        }
		if this.IsVerbose() {
			log.Printf("TapWroker.GoRev: % x\n", data[:n])
		}

		dorecv(data[:n])
	}
}

func (this *TapWroker) DoSend(data []byte) error {
	if this.IsVerbose() {
		log.Printf("TapWroker.DoSend: % x\n", data)
	}

	this.writechan <- data
	return nil
}

func (this *TapWroker) GoLoop() error {
	defer this.ifce.Close()
	for {
		select {
		case wdata := <- this.writechan:
			if _, err := this.ifce.Write(wdata); err != nil {
				log.Printf("Error|TapWroker.GoLoop: %s", err)	
			}
		}
	}
}

func (this *TapWroker) IsVerbose() bool {
    return this.verbose != 0
}