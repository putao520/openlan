package point

import (
	"context"
	"encoding/json"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"strings"
	"time"
)

type OnTcpWorker interface {
	OnIpAddr(*TcpWorker, *models.Network) error
}

type TcpWorker struct {
	On     OnTcpWorker
	Client *libol.TcpClient

	readChan  chan []byte
	writeChan chan []byte
	maxSize   int
	alias     string
	user      *models.User
	network   *models.Network
	routes    map[string]*models.Route
}

func NewTcpWorker(client *libol.TcpClient, c *config.Point, on OnTcpWorker) (t *TcpWorker) {
	t = &TcpWorker{
		On:        on,
		Client:    client,
		writeChan: make(chan []byte, 1024*10),
		maxSize:   c.IfMtu,
		user:      models.NewUser(c.Name(), c.Password()),
		network:   models.NewNetwork(c.Tenant(), c.IfAddr),
		routes:    make(map[string]*models.Route, 64),
	}
	t.user.Alias = c.Alias
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
	body, err := json.Marshal(t.user)
	if err != nil {
		libol.Error("TcpWorker.TryLogin: %s", err)
		return err
	}

	libol.Info("TcpWorker.TryLogin: %s", body)
	if err := client.SendReq("login", string(body)); err != nil {
		return err
	}
	return nil
}

func (t *TcpWorker) ReqNet(client *libol.TcpClient) error {
	body, err := json.Marshal(t.network)
	if err != nil {
		libol.Error("TcpWorker.ReqNet: %s", err)
		return err
	}

	libol.Info("TcpWorker.ReqNet: %s", body)
	if err := client.SendReq("ipaddr", string(body)); err != nil {
		return err
	}
	return nil
}

func (t *TcpWorker) onInstruct(data []byte) error {
	action, resp := libol.DecodeCmdAndParams(data)

	if action == "logi:" {
		libol.Debug("TcpWorker.onInstruct.login: %s", resp)
		if resp[:4] == "okay" {
			t.Client.SetStatus(libol.CLAUEHED)
			t.ReqNet(t.Client)
			libol.Info("TcpWorker.onInstruct.login: success")
		} else {
			t.Client.SetStatus(libol.CLUNAUTH)
			libol.Error("TcpWorker.onInstruct.login: %s", resp)
		}
	}

	if libol.IsErrorResponse(resp) {
		libol.Error("TcpWorker.onInstruct.%s: %s", action, resp)
		return nil
	}

	if action == "ipad:" {
		net := models.Network{}
		if err := json.Unmarshal([]byte(resp), &net); err != nil {
			libol.Errer("TcpWorker.onInstruct.ipaddr: Invalid json data.")
		}

		libol.Debug("TcpWorker.onInstruct.ipaddr: %s", resp)
		t.On.OnIpAddr(t, &net)

	}

	return nil
}

func (t *TcpWorker) GoRecv(ctx context.Context, doRecv func([]byte) error) {
	defer libol.Catch("TcpWorker.GoRecv")
	defer t.Close()

	libol.Info("TcpWorker.GoRev %t", t.Client.IsOk())

	for {
		if t.Client.IsTerminal() {
			return
		}

		if !t.Client.IsOk() {
			time.Sleep(30 * time.Second) // sleep 30s and release cpu.
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
			if libol.IsControl(data) {
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
	defer libol.Catch("TcpWorker.GoLoop")
	defer t.Client.Close()

	for {
		select {
		case w := <-t.writeChan:
			if t.Client.GetStatus() != libol.CLAUEHED {
				t.Client.Dropped++
				libol.Error("TcpWorker.GoLoop: dropping by unAuth")
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
	return t.user.Name, t.user.Password
}

func (t *TcpWorker) SetAuth(auth string) {
	values := strings.Split(auth, ":")
	t.user.Name = values[0]
	if len(values) > 1 {
		t.user.Password = values[1]
	}
}

func (t *TcpWorker) SetAddr(addr string) {
	t.Client.Addr = addr
}
