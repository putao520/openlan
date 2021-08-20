package api

import (
	"github.com/danieldin95/openlan/src/models"
	"github.com/danieldin95/openlan/src/olsw/store"
	"github.com/danieldin95/openlan/src/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Network struct {
}

func (h Network) Router(router *mux.Router) {
	router.HandleFunc("/api/network", h.List).Methods("GET")
	router.HandleFunc("/api/network/{id}", h.Get).Methods("GET")
}

func (h Network) List(w http.ResponseWriter, r *http.Request) {
	nets := make([]schema.Network, 0, 1024)
	for u := range store.Network.List() {
		if u == nil {
			break
		}
		nets = append(nets, models.NewNetworkSchema(u))
	}
	ResponseJson(w, nets)
}

func (h Network) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	net := store.Network.Get(vars["id"])
	if net != nil {
		ResponseJson(w, models.NewNetworkSchema(net))
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}
