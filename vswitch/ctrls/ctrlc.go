package ctrls

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/vswitch/service"
)

type CtrlC struct {
	Conn *ctl.Conn
}

func (cc *CtrlC) Register() {
	_ = service.Point.Listen.Add("ctlc", &Point{cc})
}

func (cc *CtrlC) Open(url, uuid, token string) error {
	ws := &libol.WsClient{
		Auth: libstar.Auth{
			Type:     "basic",
			Username: uuid,
			Password: token,
		},
		Url: url,
	}
	ws.Initialize()
	to, err := ws.Dial()
	if err != nil {
		return err
	}
	cc.Conn = &ctl.Conn{
		Conn: to,
	}
	return nil
}

func (cc *CtrlC) Start() {
	cc.Register()
	if cc.Conn != nil {
		cc.Conn.Open()
		cc.Conn.Start()
		cc.Conn.Send(ctl.Message{Resource: "hello", Data: "from server"})
	}
}

func (cc *CtrlC) Stop() {
	if cc.Conn != nil {
		cc.Conn.Stop()
	}
}

func (cc *CtrlC) Send(m ctl.Message) {
	if cc.Conn != nil {
		cc.Conn.Send(m)
	}
}
