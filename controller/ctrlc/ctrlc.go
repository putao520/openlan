package ctrlc

import "github.com/danieldin95/openlan-go/controller/libctrl"

type CtrlC struct {
	Conn *libctrl.Conn
}

func (cc *CtrlC) Register() {
	if cc.Conn != nil {
		cc.Conn.Listener("hello", &Hello{cc: cc})
		cc.Conn.Listener("point", &Point{cc: cc})
		cc.Conn.Listener("link", &Link{cc: cc})
		cc.Conn.Listener("neighbor", &Neighbor{cc: cc})
		cc.Conn.Listener("switch", &Switch{cc: cc})
	}
}

func (cc *CtrlC) Start() {
	cc.Register()
	if cc.Conn != nil {
		cc.Conn.Open()
		cc.Conn.Start()
		// Get all include point, link and etc.
		cc.Conn.Send(libctrl.Message{Resource: "switch"})
		cc.Conn.Send(libctrl.Message{Resource: "point"})
		cc.Conn.Send(libctrl.Message{Resource: "link"})
		cc.Conn.Send(libctrl.Message{Resource: "neighbor"})
		cc.Conn.Send(libctrl.Message{Resource: "online"})
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
