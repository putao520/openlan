package olan

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

type Ctrl struct {
}

func (w Ctrl) Router(router *mux.Router) {
	router.Handle("/olan/ctrl", websocket.Handler(w.Handle))
}

func (w Ctrl) Handle(ws *websocket.Conn) {
	defer ws.Close()

	conn := ctl.Conn{
		Conn: ws,
		Wait: libstar.NewWaitOne(1),
	}
	conn.Open()
	conn.Start()
	conn.Send(ctl.Message{Type: "hello", Data: "from server"})
	conn.Wait.Wait()
	libol.Warn("Ctrl.Handle %s exit", ws.RemoteAddr())
}
