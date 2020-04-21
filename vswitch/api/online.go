package api

import (
	"github.com/danieldin95/openlan-go/vswitch/schema"
	"github.com/danieldin95/openlan-go/vswitch/service"
	"github.com/gorilla/mux"
	"net/http"
)

type OnLine struct {
}

func (h OnLine) Router(router *mux.Router) {
	router.HandleFunc("/api/online", h.List).Methods("GET")
}

func (h OnLine) List(w http.ResponseWriter, r *http.Request) {
	nets := make([]schema.OnLine, 0, 1024)
	for u := range service.Online.List() {
		if u == nil {
			break
		}
		nets = append(nets, schema.NewOnLine(u))
	}
	ResponseJson(w, nets)
}
