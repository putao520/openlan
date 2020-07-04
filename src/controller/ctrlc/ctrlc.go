package ctrlc

import (
	"github.com/danieldin95/openlan-go/src/controller/libctrl"
	"github.com/danieldin95/openlan-go/src/switch/schema"
)

type CtrlC struct {
	Conn *libctrl.CtrlConn
}

func (cc *CtrlC) register() {
	cc.Conn.Listener("hello", &Hello{cc: cc})
	cc.Conn.Listener("point", &Point{cc: cc})
	cc.Conn.Listener("link", &Link{cc: cc})
	cc.Conn.Listener("neighbor", &Neighbor{cc: cc})
	cc.Conn.Listener("switch", &Switch{cc: cc})

	cc.Conn.Caller.Close = func(con *libctrl.CtrlConn) {
		// Clear points.
		ds := make([]string, 0, 32)
		Storager.Point.Iter(func(k string, v interface{}) {
			if p, ok := v.(*schema.Point); ok {
				if p.Switch == cc.Conn.Id {
					ds = append(ds, k)
				}
			}
		})
		for _, k := range ds {
			Storager.Point.Del(k)
		}
		// Remove switch.
		Storager.Switch.Del(cc.Conn.Id)
	}
	cc.Conn.Caller.Open = func(con *libctrl.CtrlConn) {
		// Get all include point, link and etc.
		con.Send(libctrl.Message{Resource: "switch"})
		con.Send(libctrl.Message{Resource: "point"})
		con.Send(libctrl.Message{Resource: "link"})
		con.Send(libctrl.Message{Resource: "neighbor"})
		con.Send(libctrl.Message{Resource: "online"})
	}
	cc.Conn.Caller.Ticker = func(con *libctrl.CtrlConn) {
		con.Send(libctrl.Message{Resource: "switch"})
		con.Send(libctrl.Message{Resource: "point"})
		con.Send(libctrl.Message{Resource: "link"})
		con.Send(libctrl.Message{Resource: "neighbor"})
		con.Send(libctrl.Message{Resource: "online"})
	}
}

func (cc *CtrlC) Start() {
	if cc.Conn != nil {
		cc.register()
		cc.Conn.Open()
		cc.Conn.Start()
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
