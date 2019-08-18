package endpoint

import (
	"log"
	"net"
)

type Endpoint struct {
	Hole *UdpHole
	Verbose bool
	Peers map[string]*net.UDPAddr
	Hosts map[string]*net.UDPAddr
}

func NewEndpoint(c *Config) (this *Endpoint) {
	this = &Endpoint {
		Hole: NewUdpHole(c),
		Verbose: c.Verbose,
		Peers: make(map[string]*net.UDPAddr),
		Hosts: make(map[string]*net.UDPAddr),
	}

	return
}

func (this *Endpoint) Start() {
	log.Printf("Info| Endpoint.Start")
	go this.Hole.GoKeepAlive()
}

func (this *Endpoint) Stop() {

}

