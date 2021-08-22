package api

import (
	"github.com/danieldin95/openlan/src/models"
	"github.com/danieldin95/openlan/src/olsw/store"
	"github.com/danieldin95/openlan/src/schema"
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
	for u := range store.Online.List() {
		if u == nil {
			break
		}
		nets = append(nets, models.NewOnLineSchema(u))
	}
	ResponseJson(w, nets)
}
