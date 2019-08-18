package endpoint

import (
	"log"
	"net"

	"github.com/danieldin95/openlan-go/olv2/openlanv2"
)

type Bridge struct {
	Hole *UdpHole
	Network *openlanv2.Network // <ip,port> is key.
	Hosts map[string]*openlanv2.Endpoint // MAC address is key.
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
		Hosts: make(map[string]*openlanv2.Endpoint),
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

func (this *Bridge) doRecv(raddr *net.UDPAddr, data []byte) error {
	if this.verbose {
		log.Printf("Info| Bridge.doRecv")
	}

	if openlanv2.IsInstruct(data) {
		return this.doInstruct(raddr, data)
	}

	return this.doEthernet(raddr, data)
}

func (this *Bridge) doInstruct(raddr *net.UDPAddr, data []byte) error {
	if this.verbose {
		log.Printf("Info| Bridge.doInstruct")
	}

	action := openlanv2.DecodeAction(data)
	if action == "onli:" {
		return this.doOnline(raddr, openlanv2.DecodeBody(data))
	}

	return nil
}

func (this *Bridge) doEthernet(raddr *net.UDPAddr, frame []byte) error {
	if this.verbose {
		log.Printf("Info| Bridge.doEthernet")
	}

	peer, ok := this.Network.Endpoints[raddr.String()]
	if !ok {
		//TODO learn peer by UUID.
		log.Printf("Error| Bridge.doEthernet %s not in my peers.", raddr)
		return nil
	}

	this.UpdateHost(peer, openlanv2.SrcAddr(frame))

	return this.Device.DoSend(frame)
}

func (this *Bridge) UpdateHost(peer *openlanv2.Endpoint, dst []byte) {
	_peer := this.FindHost(dst)
	if _peer != peer {
		log.Printf("Info| Bridge.UpdateHost % x change peer to %s.", dst, peer)
		this.Hosts[openlanv2.EthAddrStr(dst)] = peer
	}
}

func (this *Bridge) FindHost(dst []byte) *openlanv2.Endpoint{
	if peer, ok := this.Hosts[openlanv2.EthAddrStr(dst)]; ok {
		return peer
	}

	return nil
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

	_peer.Update(peer)
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
	peer := this.FindHost(openlanv2.DstAddr(frame))
	if peer != nil {
		return this.unicast(peer, frame)
	}
	
	return this.flood(frame)
}

func (this *Bridge) unicast(peer *openlanv2.Endpoint, frame []byte) error {
	if this.IsLocal(peer) {
		return nil
	}

	if this.verbose {
		log.Printf("Debug| Bridge.unicast to %s", peer)
	}

	if err := this.Hole.DoSend(peer.UdpAddr, peer.UUID, frame); err != nil {
		log.Printf("Error| Bridge.unicast.DoSend %s: %s", peer.UdpAddr, err)
	}

	return nil
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

		if err := this.Hole.DoSend(peer.UdpAddr, peer.UUID, frame); err != nil {
			log.Printf("Error| Bridge.flood.DoSend %s: %s", peer.UdpAddr, err)
			continue
		}
	}

	return nil
}

func (this *Bridge) IsLocal(peer *openlanv2.Endpoint) bool {
	return peer.UUID == this.Hole.UUID
}