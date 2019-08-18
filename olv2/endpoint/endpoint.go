package endpoint

import (
	"log"
	"net"
	"fmt"
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
	go this.Hole.GoAlive()
	go this.Hole.GoRecv()
}

func (this *Endpoint) Stop() {
	//TODO stop goroute.
	this.Hole.Close()
}

func (this *Endpoint) AddPeer(peer string) error {
	if _, ok := this.Peers[peer]; ok {
		return nil
	}

	raddr, err := net.ResolveUDPAddr("udp", peer)
    if err != nil {
        return err
	}

	this.Peers[peer] = raddr
	return nil
}

func (this *Endpoint) DelPeer(peer string) error {
	if _, ok := this.Peers[peer]; ok {
		delete(this.Peers, peer)
	}
	return nil
}

func (this *Endpoint) AddHost(dest []byte, peer string) error {
	dstr := fmt.Sprintf("%:x", dest)
	if _, ok := this.Hosts[dstr]; ok {
		return nil
	}

	raddr, err := net.ResolveUDPAddr("udp", peer)
    if err != nil {
        return err
	}

	this.Hosts[dstr] = raddr
	return nil
}

func (this *Endpoint) DelHost(dest []byte, peer string) error {
	dstr := fmt.Sprintf("%:x", dest)
	if _, ok := this.Hosts[dstr]; ok {
		delete(this.Hosts, dstr)
	}

	return nil
}

