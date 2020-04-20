package olan

import (
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
}
