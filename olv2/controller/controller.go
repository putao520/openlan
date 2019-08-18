package controller

import (
	"log"
	"net"
)

type Network struct {
	Name string
	Endpoints map[string]*net.UDPAddr
}

func NewNetwork(name string)(this *Network) {
	this = &Network {
		Name: name,
		Endpoints: make(map[string]*net.UDPAddr),
	}
	return
}

type Controller struct {
	Broker *UdpBroker
	Networks map[string]*Network
	//
	verbose bool
}

func NewController(c *Config) (this *Controller) {
	this = &Controller {
		Broker: NewUdpBroker(c),
		verbose: c.Verbose,
		Networks: make(map[string]*Network),
	}

	return
}

func (this *Controller) Start() {
	log.Printf("Info| Controller.Start")
	go this.Broker.GoRecv()
}

func (this *Controller) Stop() {
	//TODO stop goroute.
	this.Broker.Close()
}