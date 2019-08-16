package olv1ope

import (
	"log"
	"strings"
	"fmt"
	"errors"

	"github.com/songgao/water"
	"github.com/milosgajdos83/tenus"
	"github.com/danieldin95/openlan-go/olv1/olv1"
)

type OpeWroker struct {
	//Public variable
	Server *TcpServer
	Clients map[*olv1.TcpClient]*water.Interface
	Users map[string]*User

	//Private variable
	verbose int
	br tenus.Bridger
	ifmtu int
}

func NewOpeWroker(server *TcpServer, brname string, verbose int) (this *OpeWroker) {
	this = &OpeWroker {
		Server: server,
		Clients: make(map[*olv1.TcpClient]*water.Interface),
		Users: make(map[string]*User),
		verbose: verbose,
		br: nil,
		ifmtu: 1514,
	}

	this.createBr(brname)

	return 
}

func (this *OpeWroker) createBr(brname string) {
	addrs := strings.Split(this.Server.GetAddr(), ":")
	if len(addrs) != 2 {
		log.Printf("Error|OpeWroker.createBr: address: %s", this.Server.GetAddr())
		return
	}

	var err error
	var br tenus.Bridger

	if (brname == "") {
		brname = fmt.Sprintf("br-olan-%s", addrs[1])
		br, err = tenus.BridgeFromName(brname)
		if err != nil {
			br, err = tenus.NewBridgeWithName(brname)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		br, err = tenus.BridgeFromName(brname)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err = br.SetLinkUp(); err != nil {
		log.Printf("Error|OpeWroker.createBr: %s", err)
	}

	log.Printf("OpeWroker.createBr %s", brname)

	this.br = br
}

func (this *OpeWroker) createTap() (*water.Interface, error) {
	log.Printf("OpeWroker.createTap")	
	ifce, err := water.New(water.Config {
        DeviceType: water.TAP,
    })
    if err != nil {
		log.Printf("Error|OpeWroker.createTap: %s", err)
		return nil, err
	}
	
	link, err := tenus.NewLinkFrom(ifce.Name())
	if err != nil {
		log.Printf("Error|OpeWroker.createTap: Get ifce %s: %s", ifce.Name(), err)
		return nil, err
	}
	
	if err := link.SetLinkUp(); err != nil {
		log.Printf("Error|OpeWroker.createTap: ", err)
	}

	if err := this.br.AddSlaveIfc(link.NetInterface()); err != nil {
		log.Printf("Error|OpeWroker.createTap: Switch ifce %s: %s", ifce.Name(), err)
		return nil, err
	}

	log.Printf("OpeWroker.createTap %s", ifce.Name())	

	return ifce, nil
}

func (this *OpeWroker) Start() {
    go this.Server.GoAccept()
    go this.Server.GoLoop(this.onClient, this.onRecv, this.onClose)
}

func (this *OpeWroker) onClient(client *olv1.TcpClient) error {
	//TODO Auth it.
	log.Printf("Info|OpeWroker.onClient: %s", client)	

	ifce, err := this.createTap()
	if err != nil {
		return err
	}
	
	this.Clients[client] = ifce
	go this.GoRecv(ifce, client.SendMsg)

	return nil
}

func (this *OpeWroker) onRecv(client *olv1.TcpClient, data []byte) error {
	//TODO Hook packets such as ARP Learning.
	if this.IsVerbose() {
		log.Printf("Info|OpeWroker.onRecv: %s % x", client, data)	
	}

	ifce := this.Clients[client]
	if ifce == nil {
		return errors.New("Tap devices is nil")
	}

	if _, err := ifce.Write(data); err != nil {
		log.Printf("Error|OpeWroker.onRecv: %s", err)	
	}

	return nil
}

func (this *OpeWroker) onClose(client *olv1.TcpClient) error {
	log.Printf("Info|OpeWroker.onClose: %s", client)
	if ifce := this.Clients[client]; ifce != nil {
		ifce.Close()
		delete(this.Clients, client)
	} 
	return nil
}

func (this *OpeWroker) Close() {
	this.Server.Close()
}

func (this *OpeWroker) GoRecv(ifce *water.Interface, dorecv func([]byte)(error)) {
	log.Printf("Info|OpeWroker.GoRecv: %s", ifce.Name())	
	defer ifce.Close()
	for {
		data := make([]byte, this.ifmtu)
        n, err := ifce.Read(data)
        if err != nil {
			log.Printf("Error|OpeWroker.GoRev: %s", err)
			break
        }
		if this.IsVerbose() {
			log.Printf("OpeWroker.GoRev: % x\n", data[:n])
		}

		if err := dorecv(data[:n]); err != nil {
			log.Printf("Error|OpeWroker.GoRev: do-recv %s %s", ifce.Name(), err)
		}
	}
}

func (this *OpeWroker) IsVerbose() bool {
	return this.verbose != 0
}