package app

import (
	"encoding/json"
	"fmt"
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/models"
	"github.com/lightstar-dev/openlan-go/vswitch/api"
	"strings"
	"sync"
)

type PointAuth struct {
	Success int
	Failed  int

	ifMtu   int
	worker  api.Worker
	lock    sync.RWMutex
	clients map[*libol.TcpClient]*models.Point
}

func NewPointAuth(w api.Worker, c *config.VSwitch) (p *PointAuth) {
	p = &PointAuth{
		ifMtu:   c.IfMtu,
		worker:  w,
		clients: make(map[*libol.TcpClient]*models.Point, 1024),
	}
	return
}

func (p *PointAuth) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("PointAuth.OnFrame % x.", frame.Data)

	if libol.IsControl(frame.Data) {
		action := libol.DecodeCmd(frame.Data)
		libol.Debug("PointAuth.OnFrame.action: %s", action)

		switch action {
		case "logi=":
			if err := p.handleLogin(client, libol.DecodeParams(frame.Data)); err != nil {
				libol.Error("PointAuth.OnFrame: %s", err)
				client.SendResp("login", err.Error())
				client.Close()
				return err
			}
			client.SendResp("login", "okay.")
		}

		//If instruct is not login, continue to process.
		return nil
	}

	//Dropped all frames if not auth.
	if client.GetStatus() != libol.CLAUEHED {
		client.Dropped++
		libol.Debug("PointAuth.onRecv: %s unAuth", client.Addr)
		return libol.Errer("unAuth client.")
	}

	return nil
}

func (p *PointAuth) handleLogin(client *libol.TcpClient, data string) error {
	libol.Debug("PointAuth.handleLogin: %s", data)

	if client.GetStatus() == libol.CLAUEHED {
		libol.Warn("PointAuth.handleLogin: already auth %s", client)
		return nil
	}

	user := models.NewUser("", "")
	if err := json.Unmarshal([]byte(data), user); err != nil {
		return libol.Errer("Invalid json data.")
	}

	name := user.Name
	if user.Token != "" {
		name = user.Token
	}
	nowUser := p.worker.GetUser(name)
	if nowUser != nil {
		if nowUser.Password == user.Password {
			p.Success++
			client.SetStatus(libol.CLAUEHED)
			libol.Info("PointAuth.handleLogin: %s auth", client.Addr)
			p.onAuth(client, user)
			return nil
		}
	}

	p.Failed++
	client.SetStatus(libol.CLUNAUTH)
	return libol.Errer("Auth failed.")
}

func (p *PointAuth) onAuth(client *libol.TcpClient, user *models.User) error {
	if client.GetStatus() != libol.CLAUEHED {
		return libol.Errer("not auth.")
	}

	libol.Info("PointAuth.onAuth: %s", client.Addr)
	dev, err := p.worker.NewTap()
	if err != nil {
		return err
	}

	m := models.NewPoint(client, dev)
	m.Alias = user.Alias

	p.AddPoint(m)
	go p.GoRecv(dev, client.SendMsg)

	return nil
}

func (p *PointAuth) GoRecv(dev *models.TapDevice, doRecv func([]byte) error) {
	libol.Info("PointAuth.GoRecv: %s", dev.Name())

	defer dev.Close()
	for {
		data := make([]byte, p.ifMtu)
		n, err := dev.Read(data)
		if err != nil {
			libol.Error("PointAuth.GoRev: %s", err)
			break
		}

		libol.Debug("PointAuth.GoRev: % x\n", data[:n])
		p.worker.Send(dev, data[:n])
		if err := doRecv(data[:n]); err != nil {
			libol.Error("PointAuth.GoRev: do-recv %s %s", dev.Name(), err)
		}
	}
}

func (p *PointAuth) AddPoint(m *models.Point) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.PubPoint(m, true)
	p.clients[m.Client] = m
}

func (p *PointAuth) GetPoint(c *libol.TcpClient) *models.Point {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if m, ok := p.clients[c]; ok {
		return m
	}
	return nil
}

func (p *PointAuth) DelPoint(c *libol.TcpClient) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if m, ok := p.clients[c]; ok {
		m.Device.Close()
		p.PubPoint(m, false)
		delete(p.clients, c)
	}
}

func (p *PointAuth) ListPoint() <-chan *models.Point {
	c := make(chan *models.Point, 128)

	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()

		for _, m := range p.clients {
			c <- m
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (p *PointAuth) PubPoint(m *models.Point, isAdd bool) {
	key := fmt.Sprintf("point:%s", strings.Replace(m.Client.String(), ":", "-", -1))
	value := map[string]interface{}{
		"remote":  m.Client.String(),
		"newTime": m.Client.NewTime,
		"device":  m.Device.Name(),
		"active":  isAdd,
	}

	if r := p.worker.GetRedis(); r != nil {
		if err := r.HMSet(key, value); err != nil {
			libol.Error("PointAuth.PubPoint hset %s", err)
		}
	}
}
