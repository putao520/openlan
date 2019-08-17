package olv1cpe

import (
	"log"
	"time"
	"fmt"

	"github.com/danieldin95/openlan-go/olv1/olv1"
)

type TcpWroker struct {
	client *olv1.TcpClient
	readchan chan []byte
	writechan chan []byte
	verbose int
	maxSize int
	name string
	password string
}

func NewTcpWoker(client *olv1.TcpClient, c *Config) (this *TcpWroker) {
	this = &TcpWroker {
		client: client,
		writechan: make(chan []byte, 1024*10),
		verbose: c.Verbose,
		maxSize: c.Ifmtu,
		name: c.Name,
		password: c.Password,
	}
	this.client.SetMaxSize(this.maxSize)
	this.client.OnConnect = this.TryLogin

	return
}

func (this *TcpWroker) TryLogin(client *olv1.TcpClient) error {
	body := fmt.Sprintf(`{"name":"%s","password":"%s"}`, this.name, this.password)
	log.Printf("Info| TcpWroker.TryLogin: %s", body)
	if err := client.SendReq("login", body); err != nil {
		return err
	}
	return nil
}

func (this *TcpWroker) onInstruct(data []byte) error {
	action := olv1.DecAction(data)
	if action == "logi:" {
		resp := olv1.DecBody(data)
		log.Printf("Info| TcpWroker.onHook.login: %s", resp)
		if resp[:4] == "okay" {
			this.client.Status = olv1.CL_AUTHED
		} else {
			this.client.Status = olv1.CL_UNAUTH
		}
	}

	return nil
}

func (this *TcpWroker) GoRecv(dorecv func([]byte)(error)) {
	defer this.client.Close()
	for {
		if !this.client.IsOk() {
			time.Sleep(2 * time.Second) // sleep 2s to release cpu.
			continue
		}

		data := make([]byte, this.maxSize)
        n, err := this.client.RecvMsg(data)
        if err != nil {
			log.Printf("Error|TcpWroker.GoRev: %s", err)
			this.client.Close()
			continue
		}
		if this.IsVerbose() {
			log.Printf("TcpWroker.GoRev: % x\n", data[:n])
		}

		if n > 0 {
			data = data[:n]
			if olv1.IsInst(data) {
				this.onInstruct(data)
			} else {
				dorecv(data)
			}
		}
	}
}

func (this *TcpWroker) DoSend(data []byte) error {
	if this.IsVerbose() {
		log.Printf("TcpWroker.DoSend: % x\n", data)
	}

	this.writechan <- data
	return nil
}

func (this *TcpWroker) GoLoop() error {
	defer this.client.Close()
	for {
		select {
		case wdata := <- this.writechan:
			if this.client.Status != olv1.CL_AUTHED {
				this.client.Droped++
				if this.IsVerbose() {
					log.Printf("Error|TcpWroker.GoLoop: droping by unauth")
					continue
				}
			}

			if err := this.client.SendMsg(wdata); err != nil {
				log.Printf("Error|TcpWroker.GoLoop: %s", err)
			}
		}
	}
}

func (this *TcpWroker) IsVerbose() bool {
    return this.verbose != 0
}