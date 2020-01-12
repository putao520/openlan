package point

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"strings"
	"time"
)

type TcpWorkerListener struct {
	OnClose   func(w *TcpWorker) error
	OnSuccess func(w *TcpWorker) error
	OnIpAddr  func(w *TcpWorker, n *models.Network) error
	ReadAt    func(p []byte) error
}

type TcpWorker struct {
	Listener TcpWorkerListener
	Client   *libol.TcpClient

	writeChan chan []byte
	maxSize   int
	alias     string
	user      *models.User
	network   *models.Network
	routes    map[string]*models.Route
	allowed   bool
}

func NewTcpWorker(client *libol.TcpClient, c *config.Point) (t *TcpWorker) {
	t = &TcpWorker{
		Client:    client,
		writeChan: make(chan []byte, 1024*10),
		maxSize:   c.IfMtu,
		user:      models.NewUser(c.Name(), c.Password()),
		network:   models.NewNetwork(c.Tenant(), c.IfAddr),
		routes:    make(map[string]*models.Route, 64),
		allowed:   c.Allowed,
	}
	t.user.Alias = c.Alias

	return
}

func (t *TcpWorker) Initialize() {
	if t.Client == nil {
		return
	}

	t.Client.SetMaxSize(t.maxSize)
	t.Client.Listener = libol.TcpClientListener{
		OnConnected: func(client *libol.TcpClient) error {
			return t.TryLogin(client)
		},
		OnClose: func(client *libol.TcpClient) error {
			return nil
		},
	}
}

func (t *TcpWorker) Start(p Pointer) {
	t.Initialize()
	t.Connect()

	go t.Read()
	go t.Loop()
}

func (t *TcpWorker) Stop() {
	close(t.writeChan)
	t.Client.Terminal()
	t.Client = nil
}

func (t *TcpWorker) Close() {
	if t.Client == nil {
		return
	}

	if t.Listener.OnClose != nil {
		t.Listener.OnClose(t)
	}
	t.Client.Close()
}

func (t *TcpWorker) Connect() error {
	s := t.Client.Status()
	if s != libol.CL_INIT {
		libol.Warn("TcpWorker.Connect status %d->%d", s, libol.CL_INIT)
		t.Client.SetStatus(libol.CL_INIT)
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
	if err := client.WriteReq("login", string(body)); err != nil {
		return err
	}
	return nil
}

func (t *TcpWorker) TryNetwork(client *libol.TcpClient) error {
	body, err := json.Marshal(t.network)
	if err != nil {
		libol.Error("TcpWorker.TryNetwork: %s", err)
		return err
	}

	libol.Info("TcpWorker.TryNetwork: %s", body)
	if err := client.WriteReq("ipaddr", string(body)); err != nil {
		return err
	}
	return nil
}

func (t *TcpWorker) onInstruct(data []byte) error {
	m := libol.NewFrameMessage(data)
	if !m.IsControl() {
		return nil
	}

	action, resp := m.CmdAndParams()
	if action == "logi:" {
		libol.Debug("TcpWorker.onInstruct.login: %s", resp)
		if resp[:4] == "okay" {
			t.Client.SetStatus(libol.CL_AUEHED)
			if t.Listener.OnSuccess != nil {
				t.Listener.OnSuccess(t)
			}
			if t.allowed {
				t.TryNetwork(t.Client)
			}
			libol.Info("TcpWorker.onInstruct.login: success")
		} else {
			t.Client.SetStatus(libol.CL_UNAUTH)
			libol.Error("TcpWorker.onInstruct.login: %s", resp)
		}

		return nil
	}

	if libol.IsErrorResponse(resp) {
		libol.Error("TcpWorker.onInstruct.%s: %s", action, resp)
		return nil
	}

	if action == "ipad:" {
		net := models.Network{}
		if err := json.Unmarshal([]byte(resp), &net); err != nil {
			libol.NewErr("TcpWorker.onInstruct.ipaddr: Invalid json data.")
		}

		libol.Debug("TcpWorker.onInstruct.ipaddr: %s", resp)
		if t.Listener.OnIpAddr != nil {
			t.Listener.OnIpAddr(t, &net)
		}

	}

	return nil
}

func (t *TcpWorker) Read() {
	libol.Info("TcpWorker.Read %t", t.Client.IsOk())
	defer libol.Catch("TcpWorker.Read")

	for {
		if t.Client == nil || t.Client.IsTerminal() {
			break
		}

		if !t.Client.IsOk() {
			time.Sleep(30 * time.Second) // sleep 30s and release cpu.
			t.Connect()
			continue
		}

		data := make([]byte, t.maxSize)
		n, err := t.Client.ReadMsg(data)
		if err != nil {
			libol.Error("TcpWorker.Read: %s", err)
			t.Close()
			continue
		}

		libol.Debug("TcpWorker.Read: % x", data[:n])
		if n > 0 {
			frame := data[:n]
			if libol.IsControl(frame) {
				t.onInstruct(frame)
			} else if t.Listener.ReadAt != nil {
				t.Listener.ReadAt(frame)
			}
		}
	}

	t.Close()
	libol.Info("TcpWorker.Read exit")
}

func (t *TcpWorker) DoWrite(data []byte) error {
	libol.Debug("TcpWorker.DoWrite: % x", data)

	t.writeChan <- data

	return nil
}

func (t *TcpWorker) Loop() {
	defer libol.Catch("TcpWorker.Loop")

	for {
		w, ok := <-t.writeChan
		if !ok || t.Client == nil {
			break
		}

		if t.Client.Status() != libol.CL_AUEHED {
			t.Client.Sts.Dropped++
			libol.Error("TcpWorker.Loop: dropping by unAuth")
			continue
		}
		if err := t.Client.WriteMsg(w); err != nil {
			libol.Error("TcpWorker.Loop: %s", err)
			break
		}
	}

	t.Close()
	libol.Info("TcpWorker.Loop exit")
}

func (t *TcpWorker) Auth() (string, string) {
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

func (t *TcpWorker) SetUUID(v string) {
	t.user.UUID = v
}
