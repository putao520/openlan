package endpoint

import (
	"log"
	"time"
	"fmt"
	"strings"
	"net"
	"sync"

	"github.com/danieldin95/openlan-go/olv2/openlanv2"
)

type Peer struct {
	Name string
	Addr *net.UDPAddr
	UUID string
}

type UdpHole struct {
	Listen string
	Udp *openlanv2.UdpSocket
	Interval time.Duration
	Trigger time.Duration
	Controller string
	UUID string

	// 
	verbose bool
	name string
	password string
	peers map[string]*Peer
	rwlock sync.RWMutex
}

func SplitAuth(auth string) (string, string) {
	values := strings.Split(auth, ":")
	if len(values) == 2 {
		return values[0], values[1]
	}
	return values[0], ""
}

func NewUdpHole(c *Config) (this *UdpHole) {
	this = &UdpHole {
		Listen: c.UdpListen,
		Udp: openlanv2.NewUdpSocket(c.UdpListen, c.Verbose),
		Interval: time.Duration(c.Interval),
		Trigger: time.Duration(1),
		Controller: c.Controller,
		UUID: openlanv2.GenUUID(16),
		name: "",
		password: "",
		peers: make(map[string]*Peer),
		verbose: c.Verbose,
	}

	this.Udp.MaxSize = c.Ifmtu+ETHLEN+int(openlanv2.HSIZE)
	log.Printf("Info| NewUdpHole UUID %s", this.UUID)
	this.name, this.password = SplitAuth(c.Auth)

	if err := this.Udp.Listen(); err != nil {
		log.Printf("Error| NewUdpHole.Listen %s\n", err)
	}

	return
}


func (this *UdpHole) AddPeer(peer *Peer) error {
	if peer.UUID == this.UUID {
		return nil
	}
	if _addr := this.GetPeer(peer.Name); _addr != nil {
		return nil
	}

	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	this.peers[peer.Name] = peer
	return nil
}

func (this *UdpHole) DelPeer(name string) error {
	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	if _, ok := this.peers[name]; ok {
		delete(this.peers, name)
	}
	return nil
}

func (this *UdpHole) GetPeer(name string) (*Peer) {
	this.rwlock.RLock()
	defer this.rwlock.RUnlock()
	if peer, ok := this.peers[name]; ok {
		return peer
	}
	return nil
}

func (this *UdpHole) GoAlive() {
    log.Printf("Info| UdpHole.GoAlive to %s\n", this.Controller)
    raddr, err := net.ResolveUDPAddr("udp", this.Controller)
    if err != nil {
		log.Printf("Error| UdpHole.GoAlive to %s\n", this.Controller)
        return
	}

	this.AddPeer(&Peer {Name: this.Controller, Addr: raddr, UUID: ""})

	n := 3
	body := fmt.Sprintf(`{"name":"%s", "password":"%s", "uuid":"%s"}`, 
						this.name, this.password, this.UUID)
	m := openlanv2.NewMessage("online", body)
	for {
		this.rwlock.RLock()
		for _, peer := range this.peers {
			err := this.Udp.SendReq(peer.Addr, peer.UUID, m)
			if err != nil {
				log.Printf("Error| UdpHole.GoAlive.SendReq %s\n", err)
			}
		}
		this.rwlock.RUnlock()

		if n > 0 {
			time.Sleep(this.Trigger * time.Second)
			n--
		} else {
			time.Sleep(this.Interval * time.Second)
		}
	}
}

func (this *UdpHole) Close() error {
	return this.Udp.Close()
}

func (this *UdpHole) GoRecv(doRecv func (raddr *net.UDPAddr, data []byte) error) {
	log.Printf("Info| UdpHole.GoRecv from %s\n", this.Controller)

	for {
		addr, uuid, data, err := this.Udp.RecvMsg()
		if err != nil {
			log.Printf("Error| UdpHole.GoRecv: %s\n", err)
			return 
		}
		if this.verbose {
			log.Printf("Info| UdpHole.GoRecv from %s,%s : %d: % x..\n", 
						addr, uuid, len(data), data[:16])
		}
		if uuid != this.UUID {
			log.Printf("Erro| UdpHole.GoRecv from %s,%s not to me.\n", addr, uuid)
		}

		if err := doRecv(addr, data); err != nil {
			log.Printf("Error| UdpHole.GoRecv from %s when doRecv %s\n", addr, err)
		}
	}
}

func (this *UdpHole) DoSend(addr *net.UDPAddr, uuid string, frame []byte) error {
	if this.verbose {
		log.Printf("Debug| UdpHole.DoSend to %s,%s\n", addr, uuid)
	}

	return this.Udp.SendMsg(addr, uuid, frame)
}



