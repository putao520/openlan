package endpoint

import (
	"log"
	"net"
	"fmt"

	"github.com/danieldin95/openlan-go/olv2/openlanv2"
)

type Bridge struct {
	Hole *UdpHole
	Network *openlanv2.Network // <ip,port> is key.
	Hosts map[string]*net.UDPAddr // MAC address is key.
	Device *Device
	//
	verbose bool
}

func NewBridge(c *Config) (this *Bridge) {
	this = &Bridge {
		Hole: NewUdpHole(c),
		Device: NewDevice(c),
		verbose: c.Verbose,
		Network: openlanv2.NewNetwork("default"),
		Hosts: make(map[string]*net.UDPAddr),
	}

	return
}

func (this *Bridge) Start() {
	log.Printf("Info| Bridge.Start")
	go this.Hole.GoAlive()
	go this.Hole.GoRecv(this.doRecv)
	go this.Device.GoRecv(this.doSend)
}

func (this *Bridge) Stop() {
	log.Printf("Info| Bridge.Stop")
	//TODO stop goroute.
	this.Hole.Close()
}

func (this *Bridge) AddHost(dest []byte, peer string) error {
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

func (this *Bridge) DelHost(dest []byte, peer string) error {
	dstr := fmt.Sprintf("%:x", dest)
	if _, ok := this.Hosts[dstr]; ok {
		delete(this.Hosts, dstr)
	}

	return nil
}

func (this *Bridge) doRecv(raddr *net.UDPAddr, data []byte) error {
	if this.verbose {
		log.Printf("Info| Bridge.doRecv")
	}

	if openlanv2.IsInst(data) {
		return this.doInstruct(raddr, data)
	}

	return this.doEthernet(raddr, data)
}

func (this *Bridge) doInstruct(raddr *net.UDPAddr, data []byte) error {
	if this.verbose {
		log.Printf("Info| Bridge.doInstruct")
	}

	action := openlanv2.DecAction(data)
	if action == "onli:" {
		return this.doOnline(raddr, openlanv2.DecBody(data))
	}

	return nil
}

func (this *Bridge) doEthernet(raddr *net.UDPAddr, frame []byte) error {
	if this.verbose {
		log.Printf("Info| Bridge.doEthernet")
	}

	// TODO learn host by source.
	return this.Device.DoSend(frame)
}

func (this *Bridge) doOnline(raddr *net.UDPAddr, body string) error {
	if this.verbose {
		log.Printf("Info| Bridge.doOnline %s", body)
	}

	peer, err := openlanv2.NewEndpointFromJson(body)
	if err != nil {
		log.Printf("Error| Bridge.doOnline: %s", err)
		return err
	}

	key := peer.UdpAddr.String()
	_peer, ok := this.Network.Endpoints[key]
	if !ok {
		_peer = peer
		this.Network.AddEndpoint(key, peer)
	}

	_peer.Update()
	//TODO update time otherwise expire.

	if this.verbose {
		log.Printf("Info| Bridge.doOnline.Network %s", this.Network)
	}
	
	return nil
}

func (this *Bridge) doSend(frame []byte) error {
	if this.verbose {
		log.Printf("Info| Bridge.doSend")
	}

	return this.forward(frame)
}

func (this *Bridge) forward(frame []byte) error {
	//TODO unicast.
	return this.flood(frame)
}

func (this *Bridge) flood(frame []byte) error {
	if this.verbose {
		log.Printf("Debug| Bridge.flood")
	}

	for _, peer := range this.Network.Endpoints {
		if this.IsLocal(peer) {
			continue
		}

		if this.verbose {
			log.Printf("Debug| Bridge.flood to %s", peer)
		}

		if err := this.Hole.DoSend(peer.UdpAddr, frame); err != nil {
			log.Printf("Error| Bridge.flood.DoSend %s: %s", peer.UdpAddr, err)
			continue
		}
	}

	return nil
}

func (this *Bridge) IsLocal(peer *openlanv2.Endpoint) bool {
	if this.verbose {
		log.Printf("Debug| Bridge.IsLocal %s:%s", peer.UUID, this.Hole.UUID)
	}
	return peer.UUID == this.Hole.UUID
}