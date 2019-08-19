package controller

import (
	"log"
	"net"

	"github.com/danieldin95/openlan-go/olv2/openlanv2"
)

type Controller struct {
	Broker *UdpBroker
	Networks map[string]*openlanv2.Network
	//
	verbose bool
}

func NewController(c *Config) (this *Controller) {
	this = &Controller {
		Broker: NewUdpBroker(c),
		verbose: c.Verbose,
		Networks: make(map[string]*openlanv2.Network),
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

	if openlanv2.IsInstruct(data) {
		return this.doInstruct(raddr, data)
	}

	return this.doEthernet(raddr, data)
}

func (this *Controller) doInstruct(raddr *net.UDPAddr, data []byte) error {
	if this.verbose {
		log.Printf("Info| Controller.doInstruct")
	}

	action := openlanv2.DecodeAction(data)
	if action == "onli=" {
		return this.doOnline(raddr, openlanv2.DecodeBody(data))
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

	from, err := openlanv2.NewEndpointFromJson(body)
	if err != nil {
		log.Printf("Error| Controller.doOnline: %s", err)
		return err
	}

	from.UdpAddr = raddr
	//TODO auth it.
	net, ok := this.Networks[from.Network]
	if !ok {
		net = openlanv2.NewNetwork(from.Network)
		this.Networks[from.Network] = net
	}
	key := from.UdpAddr.String()
	_from, ok := net.Endpoints[key]
	if !ok {
		_from = from
		net.AddEndpoint(key, from)
	}

	//TODO If UDP hole is changed.
	_from.Update(from)

	//Message for announcing to peers.
	var fromm *openlanv2.Message

	frombody, err := _from.ToJson()
	if err == nil {
		fromm = openlanv2.NewMessage("online", frombody)
	}

	for uuid, _peer := range net.Endpoints {
		if this.verbose {
			log.Printf("Debug| doOnline resp: <%s>:%s to %s", uuid, _peer.UdpAddr, raddr)
		}

		//Announce to from
		body, err := _peer.ToJson()
		if err == nil {
			m := openlanv2.NewMessage("online", body)
			this.Broker.DoSend(_from.UdpAddr, _from.UUID, m)
		} else {
			log.Printf("Error| doOnline %s: %s", _peer.UdpAddr, err)
		}

		//Announce to peers.
		if fromm != nil {
			this.Broker.DoSend(_peer.UdpAddr, _peer.UUID, fromm)
		} 
	}

	return nil
}
