package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/service"
	"strings"
)

type PointAuth struct {
	success int
	failed  int

	master Master
}

func NewPointAuth(m Master, c config.VSwitch) (p *PointAuth) {
	p = &PointAuth{
		master: m,
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
				_ = client.WriteResp("login", err.Error())
				client.Close()
				return err
			}
			_ = client.WriteResp("login", "okay.")
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
	if name != "" && user.Tenant != "" { // reset username if haven't tenant.
		if !strings.Contains(name, "@") {
			name = name + "@" + user.Tenant
		}
	}
	if user.Tenant == "" { // reset tenant by username if tenant is ''.
		if strings.Contains(name, "@") {
			user.Tenant = strings.SplitN(name, "@", 2)[1]
		}
	}
	if user.Token != "" {
		name = user.Token
	}
	nowUser := service.User.Get(name)
	if nowUser != nil {
		if nowUser.Password == user.Password {
			p.success++
			client.SetStatus(libol.CL_AUEHED)
			libol.Info("PointAuth.handleLogin: %s auth", client.Addr)
			_ = p.onAuth(client, user)
			return nil
		}
	}

	p.failed++
	client.SetStatus(libol.CL_UNAUTH)
	return libol.NewErr("Auth failed.")
}

func (p *PointAuth) onAuth(client *libol.TcpClient, user *models.User) error {
	if client.Status() != libol.CL_AUEHED {
		return libol.NewErr("not auth.")
	}

	libol.Info("PointAuth.onAuth: %s", client)
	dev, err := p.master.NewTap(user.Tenant)
	if err != nil {
		return err
	}
	m := models.NewPoint(client, dev)
	m.Alias = user.Alias
	m.UUID = user.UUID
	m.Tenant = user.Tenant
	if m.UUID == "" {
		m.UUID = user.Alias
	}
	client.SetPrivate(m)
	service.Point.Add(m)
	go p.master.ReadTap(dev, client.WriteMsg)

	return nil
}

func (p *PointAuth) Stats() (success, failed int) {
	return p.success, p.failed
}
