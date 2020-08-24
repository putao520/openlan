package app

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/switch/storage"
	"strings"
)

type PointAuth struct {
	success int
	failed  int
	master  Master
}

func NewPointAuth(m Master, c config.Switch) *PointAuth {
	return &PointAuth{
		master: m,
	}
}

func (p *PointAuth) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	if libol.HasLog(libol.LOG) {
		libol.Log("PointAuth.OnFrame %s.", frame)
	}
	if frame.IsControl() {
		action, params := frame.CmdAndParams()
		libol.Debug("PointAuth.OnFrame: %s", action)
		switch action {
		case libol.LoginReq:
			if err := p.handleLogin(client, params); err != nil {
				libol.Error("PointAuth.OnFrame: %s", err)
				m := libol.NewControlFrame(libol.LoginResp, []byte(err.Error()))
				_ = client.WriteMsg(m)
				//client.Close()
				return err
			}
			m := libol.NewControlFrame(libol.LoginResp, []byte("okay"))
			_ = client.WriteMsg(m)
		}
		//If instruct is not login and already auth, continue to process.
		if client.Status() == libol.ClAuth {
			return nil
		}
	}
	//Dropped all frames if not auth.
	if client.Status() != libol.ClAuth {
		libol.Debug("PointAuth.OnFrame: %s unAuth", client.Address())
		return libol.NewErr("unAuth client.")
	}
	return nil
}

func (p *PointAuth) handleLogin(client libol.SocketClient, data []byte) error {
	libol.Debug("PointAuth.handleLogin: %s", data)

	if client.Status() == libol.ClAuth {
		libol.Warn("PointAuth.handleLogin: already auth %s", client)
		return nil
	}

	user := models.NewUser("", "")
	if err := json.Unmarshal([]byte(data), user); err != nil {
		return libol.NewErr("Invalid json data.")
	}
	// to support lower version
	if user.Network == "" {
		if strings.Contains(user.Name, "@") {
			user.Network = strings.SplitN(user.Name, "@", 2)[1]
		} else {
			user.Network = "default"
		}
	}
	name := user.Name
	if !strings.Contains(name, "@") {
		name = name + "@" + user.Network
	}

	libol.Info("PointAuth.handleLogin: %s %s on %s", client, name, user.Alias)
	nowUser := storage.User.Get(name)
	if nowUser != nil {
		if nowUser.Password == user.Password {
			p.success++
			client.SetStatus(libol.ClAuth)
			libol.Info("PointAuth.handleLogin: %s success", client)
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
	dev, err := p.master.NewTap(user.Network)
	if err != nil {
		return err
	}
	alias := strings.ToLower(user.Alias)
	libol.Info("PointAuth.onAuth: %s on %s", client, dev.Name())
	m := models.NewPoint(client, dev)
	m.User = user.Name
	m.Alias = alias
	m.UUID = user.UUID
	m.Network = user.Network
	if m.UUID == "" {
		m.UUID = alias
	}
	// free point has same uuid.
	if om := storage.Point.GetByUUID(m.UUID); om != nil {
		libol.Info("PointAuth.onAuth: %s OffClient %s", client, om.Client)
		p.master.OffClient(om.Client)
	}
	client.SetPrivate(m)
	storage.Point.Add(m)
	libol.Go(func() {
		p.master.ReadTap(dev, func(f *libol.FrameMessage) error {
			if err := client.WriteMsg(f); err != nil {
				p.master.OffClient(client)
				return err
			}
			return nil
		})
	})
	return nil
}

func (p *PointAuth) Stats() (success, failed int) {
	return p.success, p.failed
}
