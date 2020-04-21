package ctl

import "github.com/danieldin95/openlan-go/libol"

type CtrlC struct {
	Conn *Conn
}

func (cc *CtrlC) Register() {
	if cc.Conn != nil {
		cc.Conn.Listener("hello", cc.Hello)
	}
}

func (cc *CtrlC) Hello(id string, m Message) error {
	libol.Cmd("CtrlC.Hello %s", m.Data)
	return nil
}

func (cc *CtrlC) Start() {
	cc.Register()
	if cc.Conn != nil {
		cc.Conn.Open()
		cc.Conn.Start()
		// hello firstly
		cc.Conn.Send(Message{Resource: "hello", Data: "from server"})
	}
}

func (cc *CtrlC) Stop() {
	if cc.Conn != nil {
		cc.Conn.Stop()
	}
}

func (cc *CtrlC) Wait() {
	if cc.Conn != nil {
		cc.Conn.Wait.Wait()
	}
}
