package olan

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/controller/http/api"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

type UpCall struct {
}

func (w UpCall) Router(router *mux.Router) {
	router.Handle("/olan/upcall", websocket.Handler(w.Handle))
}

func (w UpCall) Handle(ws *websocket.Conn) {
	defer ws.Close()

	req := ws.Request()
	if req == nil {
		libol.Error("UpCall.Handle request is nil")
		return
	}
	id, _, _ := api.GetAuth(req)
	if id == "" {
		libol.Error("UpCall.Handle user notFound")
		return
	}
	uc := ctl.UpCallC{
		Conn: &ctl.Conn{
			Id:   id,
			Conn: ws,
			Wait: libstar.NewWaitOne(1),
		},
	}
	uc.Start()
	uc.Wait()
	libol.Warn("UpCall.Handle %s exit", ws.RemoteAddr())
}
