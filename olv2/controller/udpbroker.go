package controller

import (
	"log"

	"github.com/danieldin95/openlan-go/olv2/openlanv2"
)

type UdpBroker struct {
	Listen string
	Udp *openlanv2.UdpSocket
	UUID string
	// 
	verbose bool
}

func NewUdpBroker(c *Config) (this *UdpBroker) {
	this = &UdpBroker {
		Listen: c.UdpListen,
		Udp: openlanv2.NewUdpSocket(c.UdpListen, c.Verbose),
		verbose: c.Verbose,
		UUID: openlanv2.GenUUID(16),
	}

	if err := this.Udp.Listen(); err != nil {
		log.Printf("Error| NewUdpBroker.Listen %s\n", err)
	}

	return
}

func (this *UdpBroker) Close() error {
	return this.Udp.Close()
}

func (this *UdpBroker) GoRecv() {
	log.Printf("Info| UdpBroker.GoRecv from %s\n", this.Listen)

	for {
		r, d, err := this.Udp.RecvMsg()
		if err != nil {
			log.Printf("Error| UdpBroker.GoRecv: %s\n", err)
			return 
		}
		if this.verbose {
			log.Printf("Info| UdpBroker.GoRecv from %s: % x\n", r, d)
		}
	}
}



