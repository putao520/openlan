package olan

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/danieldin95/openlan-go/controller/ctrlc"
	"github.com/danieldin95/openlan-go/controller/http/api"
	"github.com/danieldin95/openlan-go/controller/libctrl"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

type Ctrl struct {
}

func (w Ctrl) Router(router *mux.Router) {
	router.Handle("/ctrl", websocket.Handler(w.Handle))
}

func (w Ctrl) Handle(ws *websocket.Conn) {
	defer ws.Close()

	req := ws.Request()
	if req == nil {
		libol.Error("Ctrl.Handle request is nil")
		return
	}
	id, _, _ := api.GetAuth(req)
	if id == "" {
		libol.Error("Ctrl.Handle user notFound")
		return
	}
	libol.Info("Ctrl.Handle accept %s", id)
	cc := ctrlc.CtrlC{
		Conn: &libctrl.Conn{
			Id:   id,
			Conn: ws,
			Wait: libstar.NewWaitOne(1),
		},
	}
	cc.Start()
	cc.Wait()
	libol.Warn("Ctrl.Handle closed %s", id)
}
