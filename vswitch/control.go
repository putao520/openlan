package vswitch

import (
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/libol"
)

type Control struct {
	Server      *libol.TcpServer
	Conf        *config.VSwitch
}

func NewControl(server *libol.TcpServer, c *config.VSwitch) *Control {
	w := Control{
		Server:    server,
		Conf: c,
	}

	return &w
}

func (w *Control) GetId() string {
	return w.Server.Addr
}

func (w *Control) String() string {
	return w.GetId()
}

func (w *Control) OnClient(client *libol.TcpClient) error {
	client.SetStatus(libol.CLCONNECTED)

	libol.Info("Control.onClient: %s", client.Addr)

	return nil
}

func (w *Control) OnRecv(client *libol.TcpClient, data []byte) error {
	libol.Debug("Control.onRecv: %s % x", client.Addr, data)

	//
	return nil
}

func (w *Control) OnClose(client *libol.TcpClient) error {
	libol.Info("Control.onClose: %s", client.Addr)

	return nil
}

func (w *Control) Start() {
	go w.Server.GoAccept()
	go w.Server.GoLoop(w)
}

func (w *Control) Stop() {
	libol.Info("Control.Close")

	w.Server.Close()
}
