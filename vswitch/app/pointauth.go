package app

import (
	"encoding/json"
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/vswitch/api"
	"github.com/lightstar-dev/openlan-go/vswitch/models"
	"github.com/songgao/water"
)

type PointAuth struct {
	Success int
	Failed  int

	ifMtu  int
	worker api.Worker
}

func NewPointAuth(w api.Worker, c *models.Config) (p *PointAuth) {
	p = &PointAuth{
		ifMtu:  c.IfMtu,
		worker: w,
	}
	return
}

func (p *PointAuth) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("PointAuth.OnFrame % x.", frame.Data)

	if libol.IsInst(frame.Data) {
		action := libol.DecAction(frame.Data)
		libol.Debug("PointAuth.OnFrame.action: %s", action)

		if action == "logi=" {
			if err := p.handleLogin(client, libol.DecBody(frame.Data)); err != nil {
				libol.Error("PointAuth.OnFrame: %s", err)
				client.SendResp("login", err.Error())
				client.Close()
				return err
			}
			client.SendResp("login", "okay.")
		}

		return nil
	}

	if client.GetStatus() != libol.CLAUEHED {
		client.Dropped++
		p.worker.GetServer().DrpCount++
		libol.Debug("PointAuth.onRecv: %s unauth", client.Addr)
		return libol.Errer("Unauthed client.")
	}

	return nil
}

func (p *PointAuth) handleLogin(client *libol.TcpClient, data string) error {
	libol.Debug("PointAuth.handleLogin: %s", data)

	if client.GetStatus() == libol.CLAUEHED {
		libol.Warn("PointAuth.handleLogin: already authed %s", client)
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
	_user := p.worker.GetUser(name)
	if _user != nil {
		if _user.Password == user.Password {
			p.Success++
			client.SetStatus(libol.CLAUEHED)
			libol.Info("PointAuth.handleLogin: %s Authed", client.Addr)
			p.onAuth(client)
			return nil
		}
	}

	p.Failed++
	client.SetStatus(libol.CLUNAUTH)
	return libol.Errer("Auth failed.")
}

func (p *PointAuth) onAuth(client *libol.TcpClient) error {
	if client.GetStatus() != libol.CLAUEHED {
		return libol.Errer("not authed.")
	}

	libol.Info("PointAuth.onAuth: %s", client.Addr)
	dev, err := p.worker.NewTap()
	if err != nil {
		return err
	}

	_p := models.NewPoint(client, dev)
	p.worker.AddPoint(_p)

	go p.GoRecv(dev, client.SendMsg)

	return nil
}

func (p *PointAuth) GoRecv(dev *water.Interface, doRecv func([]byte) error) {
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
		p.worker.GetServer().TxCount++
		if err := doRecv(data[:n]); err != nil {
			libol.Error("PointAuth.GoRev: do-recv %s %s", dev.Name(), err)
		}
	}
}
