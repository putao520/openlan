package ctrls

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olctl/libctrl"
	"github.com/danieldin95/openlan-go/src/olsw/storage"
	"time"
)

type CtrlC struct {
	Url      string            `json:"url"`
	Name     string            `json:"name"`
	Password string            `json:"password"`
	Conn     *libctrl.CtrlConn `json:"connection"`
	Switcher Switcher          `json:"-"`
}

func (cc *CtrlC) Register() {
	// Listen change and update.
	_ = storage.Point.Listen.Add("ctlc", &Point{cc: cc})
	_ = storage.Neighbor.Listen.Add("ctlc", &Neighbor{cc: cc})
}

func (cc *CtrlC) Handle() {
	// Handle command
	if cc.Conn == nil {
		return
	}
	cc.Conn.Listener("point", &Point{cc: cc})
	cc.Conn.Listener("neighbor", &Neighbor{cc: cc})
	cc.Conn.Listener("online", &OnLine{cc: cc})
	cc.Conn.Listener("switch", &Switch{cc: cc})
}

func (cc *CtrlC) Open() error {
	libol.Debug("CtrlC.Open %s %s", cc.Url, cc.Password)
	ws := &libol.WsClient{
		Auth: libol.Auth{
			Type:     "basic",
			Username: cc.Name,
			Password: cc.Password,
		},
		Url: cc.Url,
	}
	ws.Initialize()
	to, err := ws.Dial()
	if err != nil {
		return err
	}
	cc.Conn = &libctrl.CtrlConn{
		Conn:    to,
		Wait:    libol.NewWaitOne(1),
		Timeout: 2 * time.Minute,
		Id:      cc.Url,
	}
	return nil
}

func (cc *CtrlC) Start() {
	if Ctrl.Url == "" {
		libol.Warn("CtrlC.Star Url is nil")
		return
	}
	cc.Register()
	for {
		_ = cc.Open()
		if cc.Conn == nil {
			time.Sleep(15 * time.Second)
			continue
		}
		cc.Handle()
		// Start it
		cc.Conn.Open()
		cc.Conn.Start()
		// Wait until it stopped.
		cc.Wait()
		cc.Conn = nil
	}
}

func (cc *CtrlC) Stop() {
	if cc.Conn != nil {
		cc.Conn.Stop()
	}
}

func (cc *CtrlC) Send(m libctrl.Message) {
	if cc.Conn != nil {
		cc.Conn.Send(m)
	}
}

func (cc *CtrlC) Wait() {
	if cc.Conn != nil {
		cc.Conn.Wait.Wait()
	}
}

var Ctrl = &CtrlC{}

func Load(path string) {
	if err := libol.FileExist(path); err != nil {
		return
	}
	if err := libol.UnmarshalLoad(Ctrl, path); err != nil {
		libol.Warn("ctrls.Load: %s", err)
		return
	}
}
