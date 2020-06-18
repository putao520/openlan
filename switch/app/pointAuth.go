package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/switch/storage"
	"strings"
)

type PointAuth struct {
	success int
	failed  int
	master  Master
}

func NewPointAuth(m Master, c config.Switch) (p *PointAuth) {
	p = &PointAuth{
		master: m,
	}
	return
}

func (p *PointAuth) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	if libol.HasLog(libol.LOG) {
		libol.Log("PointAuth.OnFrame %s.", frame)
	}
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
		//If instruct is not login and already auth, continue to process.
		if client.Status() == libol.ClAuth {
			return nil
		}
	}
	//Dropped all frames if not auth.
	if client.Status() != libol.ClAuth {
		libol.Debug("PointAuth.OnFrame: %s unAuth", client.Addr())
		return libol.NewErr("unAuth client.")
	}
	return nil
}

func (p *PointAuth) handleLogin(client libol.SocketClient, data string) error {
	libol.Debug("PointAuth.handleLogin: %s", data)

	if client.Status() == libol.ClAuth {
		libol.Warn("PointAuth.handleLogin: already auth %s", client)
		return nil
	}

	user := models.NewUser("", "")
	if err := json.Unmarshal([]byte(data), user); err != nil {
		return libol.NewErr("Invalid json data.")
	}

	name := user.Name
	if name != "" && user.Network != "" { // reset username if haven't tenant.
		if !strings.Contains(name, "@") {
			name = name + "@" + user.Network
		}
	}
	if user.Network == "" { // reset tenant by username if tenant is ''.
		if strings.Contains(name, "@") {
			user.Network = strings.SplitN(name, "@", 2)[1]
		}
	}
	if user.Token != "" {
		name = user.Token
	}
	libol.Info("PointAuth.handleLogin: %s on %s", name, user.Alias)
	nowUser := storage.User.Get(name)
	if nowUser != nil {
		if nowUser.Password == user.Password {
			p.success++
			client.SetStatus(libol.ClAuth)
			libol.Info("PointAuth.handleLogin: %s auth", client.Addr())
			_ = p.onAuth(client, user)
			return nil
		}
	}
	p.failed++
	client.SetStatus(libol.ClUnAuth)
	return libol.NewErr("Auth failed.")
}

func (p *PointAuth) onAuth(client libol.SocketClient, user *models.User) error {
	if client.Status() != libol.ClAuth {
		return libol.NewErr("not auth.")
	}

	libol.Info("PointAuth.onAuth: %s", client)
	d, err := p.master.NewTap(user.Network)
	if err != nil {
		return err
	}
	m := models.NewPoint(client, d)
	m.Alias = user.Alias
	m.UUID = user.UUID
	m.Network = user.Network
	if m.UUID == "" {
		m.UUID = user.Alias
	}
	// free point has same uuid.
	if om := storage.Point.GetByUUID(m.UUID); om != nil {
		p.master.OffClient(om.Client)
	}
	client.SetPrivate(m)
	storage.Point.Add(m)
	libol.Go(func() { p.master.ReadTap(d, client.WriteMsg) })
	return nil
}

func (p *PointAuth) Stats() (success, failed int) {
	return p.success, p.failed
}
