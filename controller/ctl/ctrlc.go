package ctl

type CtrlC struct {
	Conn *Conn
}

func (cc *CtrlC) Register() {
	if cc.Conn != nil {
		cc.Conn.Listener("hello", &Hello{cc})
		cc.Conn.Listener("point", &Point{cc})
		cc.Conn.Listener("link", &Link{cc})
		cc.Conn.Listener("neighbor", &Neighbor{cc})
	}
}

func (cc *CtrlC) Start() {
	cc.Register()
	if cc.Conn != nil {
		cc.Conn.Open()
		cc.Conn.Start()
		// Get all include point, link and etc.
		cc.Conn.Send(Message{Resource: "point"})
		cc.Conn.Send(Message{Resource: "link"})
		cc.Conn.Send(Message{Resource: "neighbor"})
		cc.Conn.Send(Message{Resource: "online"})
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
