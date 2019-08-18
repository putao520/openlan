package endpoint

import (
	"github.com/danieldin95/openlan-go/olv2/openlanv2"
)

type Endpoint struct {
	Udp *openlanv2.UdpSocket
	Verbose int
	Peers map[string]*net.UDPAddr
	Hosts map[[]byte]*net.UDPAddr
}

func NewEndpoint(c *Config) (this *Endpoint) {
	this = &Endpoint {
		Udp: NewTcpSocket(c.UdpListen, c.Verbose),
		Verbose: c.Verbose,
		Peers: make(map[string]*net.UDPAddr),
		Hosts: make(map[[]byte]*net.UDPAddr),
	}
	if err := this.Listen(); err != nil {
		log.Printf("Error| NewEndpoint.Listen %s\n", err)
	}

	return
}


