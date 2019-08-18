package controller

import (
	"log"
	"net"

	"github.com/danieldin95/openlan-go/olv2/openlanv2"
)

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
	go this.Broker.GoRecv(this.doRecv)
}

func (this *Controller) Stop() {
	//TODO stop goroute.
	this.Broker.Close()
}

func (this *Controller) doRecv(raddr *net.UDPAddr, data []byte) error {
	if this.verbose {
		log.Printf("Info| Controller.doRecv")
	}

	if openlanv2.IsInst(data) {
		return this.doInstruct(raddr, data)
	}

	return this.doEthernet(raddr, data)
}

func (this *Controller) doInstruct(raddr *net.UDPAddr, data []byte) error {
	if this.verbose {
		log.Printf("Info| Controller.doInstruct")
	}

	action := openlanv2.DecAction(data)
	if action == "onli=" {
		return this.doOnline(raddr, openlanv2.DecBody(data))
	}

	return nil
}

func (this *Controller) doEthernet(raddr *net.UDPAddr, data []byte) error {
	if this.verbose {
		log.Printf("Info| Controller.doEthernet")
	}

	return nil
}

func (this *Controller) doOnline(raddr *net.UDPAddr, body string) error {
	if this.verbose {
		log.Printf("Info| Controller.doOnline")
	}

	point, err := NewEndpointFromJson(body)
	if err != nil {
		log.Printf("Error| Controller.doOnline: %s", err)
		return err
	}

	//TODO auth it.
	net, ok := this.Networks[point.Network]
	if !ok {
		net = NewNetwork(point.Network)
		this.Networks[point.Network] = net
	}
	_point, ok := net.Endpoints[point.UUID]
	if !ok {
		_point = point
		net.AddEndpoint(point)
	}

	//TODO If UDP hole is changed.
	_point.UdpAddr = raddr

	for uuid, peer := range net.Endpoints {
		if this.verbose {
			log.Printf("Debug| doOnline resp: %s:%s to %s", uuid, peer.UdpAddr, raddr)
		}

		body, err := peer.ToJson()
		if err == nil {
			this.Broker.DoSend(raddr, "online", body)
		} else {
			log.Printf("Error| doOnline %s: %s", peer.UdpAddr, err)
		}
	}

	return nil
}
