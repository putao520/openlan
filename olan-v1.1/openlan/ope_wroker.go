package openlanv1

import (
	"log"
	"strings"
	"fmt"
	"net"
	"errors"

	"github.com/songgao/water"
	"github.com/milosgajdos83/tenus"
)

type OpeWroker struct {
	Server *TcpServer
	verbose int
	br tenus.Bridger
	clients map[*TcpClient]*water.Interface
	ifmtu int
}

func NewOpeWroker(server *TcpServer, brname string, verbose int) (this *OpeWroker) {
	this = &OpeWroker {
		Server: server,
		verbose: verbose,
		br: nil,
		ifmtu: 1514,
		clients: make(map[*TcpClient]*water.Interface),
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
	
	link, err := net.InterfaceByName(ifce.Name())
	if err != nil {
		log.Printf("Error|OpeWroker.createTap: Get ifce %s", ifce.Name(), err)
		return nil, err
	}
	
	if err := this.br.AddSlaveIfc(link); err != nil {
		log.Printf("Error|OpeWroker.createTap: Switch ifce %s", ifce.Name(), err)
		return nil, err
	}

	log.Printf("OpeWroker.createTap %s", ifce.Name())	

	return ifce, nil
}

func (this *OpeWroker) Start() {
    go this.Server.GoAccept()
    go this.Server.GoLoop(this.onClient, this.onRecv, this.onClose)
}

func (this *OpeWroker) onClient(client *TcpClient) error {
	log.Printf("Info|OpeWroker.onClient: %s", client)	

	ifce, err := this.createTap()
	if err != nil {
		return err
	}
	
	this.clients[client] = ifce
	go this.GoRecv(ifce, client.SendMsg)

	return nil
}

func (this *OpeWroker) onRecv(client *TcpClient, data []byte) error {
	if this.IsVerbose() {
		log.Printf("Info|OpeWroker.onRecv: %s % x", client, data)	
	}

	ifce := this.clients[client]
	if ifce == nil {
		return errors.New("Tap devices is nil")
	}

	if _, err := ifce.Write(data); err != nil {
		log.Printf("Error|OpeWroker.onRecv: %s", err)	
	}

	return nil
}

func (this *OpeWroker) onClose(client *TcpClient) error {
	log.Printf("Info|OpeWroker.onClose: %s", client)
	if ifce := this.clients[client]; ifce != nil {
		ifce.Close()
		delete(this.clients, client)
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