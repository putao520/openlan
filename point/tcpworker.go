package point

import (
	"fmt"
	"strings"
	"time"

	"github.com/lightstar-dev/openlan-go/libol"
)

type TcpWorker struct {
	client    *libol.TcpClient
	readchan  chan []byte
	writechan chan []byte
	maxSize   int
	name      string
	password  string
}

func NewTcpWorker(client *libol.TcpClient, c *Config) (t *TcpWorker) {
	t = &TcpWorker{
		client:    client,
		writechan: make(chan []byte, 1024*10),
		maxSize:   c.Ifmtu,
		name:      c.Name(),
		password:  c.Password(),
	}
	t.client.SetMaxSize(t.maxSize)
	t.client.OnConnected(t.TryLogin)

	return
}

func (t *TcpWorker) Close() {
	t.client.Close()
}

func (t *TcpWorker) Connect() error {
	return t.client.Connect()
}

func (t *TcpWorker) TryLogin(client *libol.TcpClient) error {
	body := fmt.Sprintf(`{"name":"%s","password":"%s"}`, t.name, t.password)
	libol.Info("TcpWorker.TryLogin: %s", body)
	if err := client.SendReq("login", body); err != nil {
		return err
	}
	return nil
}

func (t *TcpWorker) onInstruct(data []byte) error {
	action := libol.DecAction(data)
	if action == "logi:" {
		resp := libol.DecBody(data)
		libol.Info("TcpWorker.onHook.login: %s", resp)
		if resp[:4] == "okay" {
			t.client.Status = libol.CL_AUTHED
		} else {
			t.client.Status = libol.CL_UNAUTH
		}
	}

	return nil
}

func (t *TcpWorker) GoRecv(doRecv func([]byte) error) {
	//TODO catch panic
	libol.Debug("TcpWorker.GoRev %s", t.client.IsOk())

	defer t.client.Close()
	for {
		if !t.client.IsOk() {
			time.Sleep(2 * time.Second) // sleep 2s and release cpu.
			continue
		}

		data := make([]byte, t.maxSize)
		n, err := t.client.RecvMsg(data)
		if err != nil {
			libol.Error("TcpWorker.GoRev: %s", err)
			t.client.Close()
			continue
		}

		libol.Debug("TcpWorker.GoRev: % x", data[:n])
		if n > 0 {
			data = data[:n]
			if libol.IsInst(data) {
				t.onInstruct(data)
			} else {
				doRecv(data)
			}
		}
	}
}

func (t *TcpWorker) DoSend(data []byte) error {
	libol.Debug("TcpWorker.DoSend: % x", data)

	t.writechan <- data

	return nil
}

func (t *TcpWorker) GoLoop() error {
	defer t.client.Close()
	for {
		select {
		case wdata := <-t.writechan:
			if t.client.Status != libol.CL_AUTHED {
				t.client.Droped++
				libol.Error("TcpWorker.GoLoop: droping by unauth")
				continue
			}

			if err := t.client.SendMsg(wdata); err != nil {
				libol.Error("TcpWorker.GoLoop: %s", err)
			}
		}
	}
}

func (t *TcpWorker) SetAuth(auth string) {
	values := strings.Split(auth, ":")
	t.name = values[0]
	if len(values) > 1 {
		t.password = values[1]
	}
}

func (t *TcpWorker) SetAddr(addr string) {
	t.client.Addr = addr
}
