package endpoint

import (
	"log"
	"time"
	"fmt"
	"strings"
	"net"

	"github.com/danieldin95/openlan-go/olv2/openlanv2"
)

type UdpHole struct {
	Udp *openlanv2.UdpSocket
	Interval time.Duration
	Controller string
	UUID string
	// 
	verbose bool
	name string
	password string
}

func SplitAuth(auth string) (string, string){
	values := strings.Split(auth, ":")
	if len(values) == 2 {
		return values[0], values[1]
	}
	return values[0], ""
}

func NewUdpHole(c *Config) (this *UdpHole) {
	this = &UdpHole {
		Udp: openlanv2.NewUdpSocket(c.UdpListen, c.Verbose),
		verbose: c.Verbose,
		Interval: 5,
		Controller: c.Controller,
		name: "",
		password: "",
		UUID: openlanv2.GenUUID(16),
	}

	this.name, this.password = SplitAuth(c.Auth)

	if err := this.Udp.Listen(); err != nil {
		log.Printf("Error| NewUdpHole.Listen %s\n", err)
	}

	return
}

func (this *UdpHole) GoAlive() {
    log.Printf("Info| UdpHole.GoAlive to %s\n", this.Controller)
    raddr, err := net.ResolveUDPAddr("udp", this.Controller)
    if err != nil {
		log.Printf("Error| UdpHole.GoAlive to %s\n", this.Controller)
        return
	}
	
	body := fmt.Sprintf(`{"name":"%s", "password":"%s", "uuid":"%s"}`, 
						this.name, this.password, this.UUID)
	for {
		err := this.Udp.SendReq(raddr, "online", body)
		if err != nil {
			log.Printf("Error| UdpHole.GoAlive.SendReq %s\n", err)
		}

		time.Sleep(this.Interval * time.Second)
	}
}

func (this *UdpHole) Close() error {
	return this.Udp.Close()
}

func (this *UdpHole) GoRecv() {
	log.Printf("Info| UdpHole.GoRecv from %s\n", this.Controller)

	for {
		r, d, err := this.Udp.RecvMsg()
		if err != nil {
			log.Printf("Error| UdpHole.GoRecv: %s\n", err)
			return 
		}
		if this.verbose {
			log.Printf("Info| UdpHole.GoRecv from %s : % x\n", r, d)
		}
	}
}



