package vswitch

import (
	"encoding/json"
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/songgao/water"
)

type Point struct {
	Client *libol.TcpClient
	Device *water.Interface
}

func NewPoint(c *libol.TcpClient, d *water.Interface) (w *Point) {
	w = &Point{
		Client: c,
		Device: d,
	}

	return
}

type PointAuth struct {
	ifMtu  int
	worker *Worker
}

func NewPointAuth(worker *Worker, c *Config) (w *PointAuth) {
	w = &PointAuth{
		ifMtu:  c.IfMtu,
		worker: worker,
	}
	return
}

func (w *PointAuth) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("PointAuth.OnFrame % x.", frame.Data)

	if libol.IsInst(frame.Data) {
		action := libol.DecAction(frame.Data)
		libol.Debug("PointAuth.OnFrame.action: %s", action)

		if action == "logi=" {
			if err := w.handleLogin(client, libol.DecBody(frame.Data)); err != nil {
				libol.Error("PointAuth.OnFrame: %s", err)
				client.SendResp("login", err.Error())
				client.Close()
				return err
			}
			client.SendResp("login", "okay.")
		}

		return nil
	}

	if client.GetStatus() != libol.ClAuthed {
		client.Dropped++
		libol.Debug("PointAuth.onRecv: %s unauth", client.Addr)
		return libol.Errer("Unauthed client.")
	}

	return nil
}

func (w *PointAuth) handleLogin(client *libol.TcpClient, data string) error {
	libol.Debug("PointAuth.handleLogin: %s", data)

	if client.GetStatus() == libol.ClAuthed {
		libol.Warn("PointAuth.handleLogin: already authed %s", client)
		return nil
	}

	user := NewUser("", "")
	if err := json.Unmarshal([]byte(data), user); err != nil {
		return libol.Errer("Invalid json data.")
	}

	name := user.Name
	if user.Token != "" {
		name = user.Token
	}
	_user := w.worker.GetUser(name)
	if _user != nil {
		if _user.Password == user.Password {
			client.SetStatus(libol.ClAuthed)
			libol.Info("PointAuth.handleLogin: %s Authed", client.Addr)
			w.onAuth(client)
			return nil
		}

		client.SetStatus(libol.ClUnauth)
	}

	return libol.Errer("Auth failed.")
}

func (w *PointAuth) onAuth(client *libol.TcpClient) error {
	if client.GetStatus() != libol.ClAuthed {
		return libol.Errer("not authed.")
	}

	libol.Info("PointAuth.onAuth: %s", client.Addr)
	dev, err := w.worker.NewTap()
	if err != nil {
		return err
	}

	p := NewPoint(client, dev)
	w.worker.AddPoint(p)

	go w.GoRecv(dev, client.SendMsg)

	return nil
}

func (w *PointAuth) GoRecv(dev *water.Interface, doRecv func([]byte) error) {
	libol.Info("PointAuth.GoRecv: %s", dev.Name())

	defer dev.Close()
	for {
		data := make([]byte, w.ifMtu)
		n, err := dev.Read(data)
		if err != nil {
			libol.Error("PointAuth.GoRev: %s", err)
			break
		}

		libol.Debug("PointAuth.GoRev: % x\n", data[:n])
		if err := doRecv(data[:n]); err != nil {
			libol.Error("PointAuth.GoRev: do-recv %s %s", dev.Name(), err)
		}
	}
}

type WithRequest struct {
	worker *Worker
}

func NewWithRequest(worker *Worker, c *Config) (w *WithRequest) {
	w = &WithRequest{
		worker: worker,
	}
	return
}

func (w *WithRequest) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
	libol.Debug("WithRequest.OnFrame % x.", frame.Data)

	if libol.IsInst(frame.Data) {
		action, body := libol.DecActionBody(frame.Data)
		libol.Debug("WithRequest.OnFrame.action: %s %s", action, body)

		if action == "neig=" {
			//TODO
		}
	}

	return nil
}
