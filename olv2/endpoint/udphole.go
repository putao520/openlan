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
	Verbose bool
	Interval time.Duration
	Controller string
	// 
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
		Verbose: c.Verbose,
		Interval: 5,
		Controller: c.Controller,
		name: "",
		password: "",
	}

	this.name, this.password = SplitAuth(c.Auth)

	if err := this.Udp.Listen(); err != nil {
		log.Printf("Error| NewEndpoint.Listen %s\n", err)
	}

	return
}

func (this *UdpHole) GoKeepAlive() {
    log.Printf("Info| UdpHole.GoKeepAlive to %s\n", this.Controller)
    raddr, err := net.ResolveUDPAddr("udp", this.Controller)
    if err != nil {
		log.Printf("Error| UdpHole.GoKeepAlive to %s\n", this.Controller)
        return
	}
	
	body := fmt.Sprintf(`{"name":"%s", "password":"%s"}`, this.name, this.password)
	for {
		err := this.Udp.SendReq(raddr, "online", body)
		if err != nil {
			log.Printf("Error| UdpHole.GoKeepAlive.SendReq %s\n", err)
		}

		time.Sleep(this.Interval * time.Second)
	}
}



