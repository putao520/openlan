package point

import (
	"fmt"
	"strings"
	"time"

	"github.com/lightstar-dev/openlan-go/libol"
)

type TcpWroker struct {
	client    *libol.TcpClient
	readchan  chan []byte
	writechan chan []byte
	maxSize   int
	name      string
	password  string
}

func NewTcpWoker(client *libol.TcpClient, c *Config) (this *TcpWroker) {
	this = &TcpWroker{
		client:    client,
		writechan: make(chan []byte, 1024*10),
		maxSize:   c.Ifmtu,
		name:      c.Name(),
		password:  c.Password(),
	}
	this.client.SetMaxSize(this.maxSize)
	this.client.OnConnected(this.TryLogin)

	return
}

func (this *TcpWroker) Close() {
	this.client.Close()
}

func (this *TcpWroker) Connect() error {
	return this.client.Connect()
}

func (this *TcpWroker) TryLogin(client *libol.TcpClient) error {
	body := fmt.Sprintf(`{"name":"%s","password":"%s"}`, this.name, this.password)
	libol.Info("TcpWroker.TryLogin: %s", body)
	if err := client.SendReq("login", body); err != nil {
		return err
	}
	return nil
}

func (this *TcpWroker) onInstruct(data []byte) error {
	action := libol.DecAction(data)
	if action == "logi:" {
		resp := libol.DecBody(data)
		libol.Info("TcpWroker.onHook.login: %s", resp)
		if resp[:4] == "okay" {
			this.client.Status = libol.CL_AUTHED
		} else {
			this.client.Status = libol.CL_UNAUTH
		}
	}

	return nil
}

func (this *TcpWroker) GoRecv(doRecv func([]byte) error) {
	libol.Debug("TcpWroker.GoRev %s\n", this.client.IsOk())

	defer this.client.Close()
	for {
		if !this.client.IsOk() {
			time.Sleep(2 * time.Second) // sleep 2s to release cpu.
			continue
		}

		data := make([]byte, this.maxSize)
		n, err := this.client.RecvMsg(data)
		if err != nil {
			libol.Error("TcpWroker.GoRev: %s", err)
			this.client.Close()
			continue
		}

		libol.Debug("TcpWroker.GoRev: % x\n", data[:n])
		if n > 0 {
			data = data[:n]
			if libol.IsInst(data) {
				this.onInstruct(data)
			} else {
				doRecv(data)
			}
		}
	}
}

func (this *TcpWroker) DoSend(data []byte) error {
	libol.Debug("TcpWroker.DoSend: % x\n", data)

	this.writechan <- data

	return nil
}

func (this *TcpWroker) GoLoop() error {
	defer this.client.Close()
	for {
		select {
		case wdata := <-this.writechan:
			if this.client.Status != libol.CL_AUTHED {
				this.client.Droped++
				libol.Error("TcpWroker.GoLoop: droping by unauth")
				continue
			}

			if err := this.client.SendMsg(wdata); err != nil {
				libol.Error("TcpWroker.GoLoop: %s", err)
			}
		}
	}
}

func (this *TcpWroker) SetAuth(auth string) {
	values := strings.Split(auth, ":")
	this.name = values[0]
	if len(values) > 1 {
		this.password = values[1]
	}
}

func (this *TcpWroker) SetAddr(addr string) {
	this.client.Addr = addr
}
