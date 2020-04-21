package ctl

import "github.com/danieldin95/openlan-go/libol"

type UpCallC struct {
	Conn *Conn
}

func (uc *UpCallC) Register() {
	if uc.Conn != nil {
		uc.Conn.Listener("hello", uc.Hello)
		uc.Conn.Listener("system", uc.System)
		uc.Conn.Listener("point", uc.Point)
		uc.Conn.Listener("link", uc.Link)
		uc.Conn.Listener("neighbor", uc.Neighbor)
	}
}

func (uc *UpCallC) Hello(id string, m Message) error {
	libol.Cmd("UpCallC.Hello %s", m.Data)
	return nil
}

func (uc *UpCallC) System(id string, m Message) error {
	libol.Cmd("UpCallC.System %s", m.Data)
	return nil
}

func (uc *UpCallC) Point(id string, m Message) error {
	libol.Cmd("UpCallC.Point %s", m.Data)
	return nil
}

func (uc *UpCallC) Link(id string, m Message) error {
	libol.Cmd("UpCallC.Link %s", m.Data)
	return nil
}

func (uc *UpCallC) Neighbor(id string, m Message) error {
	libol.Cmd("UpCallC.Neighbor %s", m.Data)
	return nil
}

func (uc *UpCallC) Start() {
	uc.Register()
	if uc.Conn != nil {
		uc.Conn.Open()
		uc.Conn.Start()
		// hello firstly
		uc.Conn.Send(Message{Resource: "hello", Data: "from server"})
	}
}

func (uc *UpCallC) Stop() {
	if uc.Conn != nil {
		uc.Conn.Stop()
	}
}

func (uc *UpCallC) Wait() {
	if uc.Conn != nil {
		uc.Conn.Wait.Wait()
	}
}
