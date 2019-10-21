package point

import (
	"context"
	"fmt"
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/libol"
	"strings"
	"time"
)

type TcpWorker struct {
	Client *libol.TcpClient

	readChan  chan []byte
	writeChan chan []byte
	maxSize   int
	name      string
	password  string
}

func NewTcpWorker(client *libol.TcpClient, c *config.Point) (t *TcpWorker) {
	t = &TcpWorker{
		Client:    client,
		writeChan: make(chan []byte, 1024*10),
		maxSize:   c.IfMtu,
		name:      c.Name(),
		password:  c.Password(),
	}
	t.Client.SetMaxSize(t.maxSize)
	t.Client.OnConnected(t.TryLogin)

	return
}

func (t *TcpWorker) Stop() {
	t.Client.Terminal()
}

func (t *TcpWorker) Close() {
	t.Client.Close()
}

func (t *TcpWorker) Connect() error {
	s := t.Client.GetStatus()
	if s != libol.CLINIT {
		libol.Warn("TcpWorker.Connect status %d->%d", s, libol.CLINIT)
		t.Client.SetStatus(libol.CLINIT)
	}

	if err := t.Client.Connect(); err != nil {
		libol.Error("TcpWorker.Connect %s", err)
		return err
	}
	return nil
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
			t.Client.SetStatus(libol.CLAUEHED)
		} else {
			t.Client.SetStatus(libol.CLUNAUTH)
		}
	}

	return nil
}

func (t *TcpWorker) GoRecv(ctx context.Context, doRecv func([]byte) error) {
	defer libol.Catch("TcpWroker.GoRecv")
	defer t.Close()

	libol.Info("TcpWorker.GoRev %t", t.Client.IsOk())

	for {
		if t.Client.IsTerminal() {
			return
		}

		if !t.Client.IsOk() {
			time.Sleep(60 * time.Second) // sleep 30s and release cpu.
			t.Connect()
			continue
		}

		data := make([]byte, t.maxSize)
		n, err := t.Client.RecvMsg(data)
		if err != nil {
			libol.Error("TcpWorker.GoRev: %s", err)
			t.Client.Close()
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

	t.writeChan <- data

	return nil
}

func (t *TcpWorker) GoLoop(ctx context.Context) {
	defer libol.Catch("TcpWroker.GoLoop")
	defer t.Client.Close()

	for {
		select {
		case w := <-t.writeChan:
			if t.Client.GetStatus() != libol.CLAUEHED {
				t.Client.Dropped++
				libol.Error("TcpWorker.GoLoop: droping by unauth")
				continue
			}

			if err := t.Client.SendMsg(w); err != nil {
				libol.Error("TcpWorker.GoLoop: %s", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
func (t *TcpWorker) GetAuth() (string, string) {
	return t.name, t.password
}

func (t *TcpWorker) SetAuth(auth string) {
	values := strings.Split(auth, ":")
	t.name = values[0]
	if len(values) > 1 {
		t.password = values[1]
	}
}

func (t *TcpWorker) SetAddr(addr string) {
	t.Client.Addr = addr
}
