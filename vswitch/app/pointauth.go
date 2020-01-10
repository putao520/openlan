package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/service"
)

type PointAuth struct {
	Success int
	Failed  int

	ifMtu  int
	worker Worker
}

func NewPointAuth(w Worker, c *config.VSwitch) (p *PointAuth) {
	p = &PointAuth{
		ifMtu:  c.IfMtu,
		worker: w,
	}
	return
}

func (p *PointAuth) OnFrame(client *libol.TcpClient, frame *libol.FrameMessage) error {
	libol.Debug("PointAuth.OnFrame %s.", frame)

	if frame.IsControl() {
		action, params := frame.CmdAndParams()
		libol.Debug("PointAuth.OnFrame: %s", action)

		switch action {
		case "logi=":
			if err := p.handleLogin(client, params); err != nil {
				libol.Error("PointAuth.OnFrame: %s", err)
				client.WriteResp("login", err.Error())
				client.Close()
				return err
			}
			client.WriteResp("login", "okay.")
		}

		//If instruct is not login, continue to process.
		return nil
	}

	//Dropped all frames if not auth.
	if client.Status() != libol.CL_AUEHED {
		client.Sts.Dropped++
		libol.Debug("PointAuth.onRead: %s unAuth", client.Addr)
		return libol.NewErr("unAuth client.")
	}

	return nil
}

func (p *PointAuth) handleLogin(client *libol.TcpClient, data string) error {
	libol.Debug("PointAuth.handleLogin: %s", data)

	if client.Status() == libol.CL_AUEHED {
		libol.Warn("PointAuth.handleLogin: already auth %s", client)
		return nil
	}

	user := models.NewUser("", "")
	if err := json.Unmarshal([]byte(data), user); err != nil {
		return libol.NewErr("Invalid json data.")
	}

	name := user.Name
	if user.Token != "" {
		name = user.Token
	}
	nowUser := service.User.Get(name)
	if nowUser != nil {
		if nowUser.Password == user.Password {
			p.Success++
			client.SetStatus(libol.CL_AUEHED)
			libol.Info("PointAuth.handleLogin: %s auth", client.Addr)
			p.onAuth(client, user)
			return nil
		}
	}

	p.Failed++
	client.SetStatus(libol.CL_UNAUTH)
	return libol.NewErr("Auth failed.")
}

func (p *PointAuth) onAuth(client *libol.TcpClient, user *models.User) error {
	if client.Status() != libol.CL_AUEHED {
		return libol.NewErr("not auth.")
	}

	libol.Info("PointAuth.onAuth: %s", client.Addr)
	dev, err := p.worker.NewTap()
	if err != nil {
		return err
	}

	m := models.NewPoint(client, dev)
	m.Alias = user.Alias

	service.Point.Add(m)
	go p.worker.ReadTap(dev, client.WriteMsg)

	return nil
}