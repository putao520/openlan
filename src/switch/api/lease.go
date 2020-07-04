package api

import (
	"github.com/danieldin95/openlan-go/src/switch/schema"
	"github.com/danieldin95/openlan-go/src/switch/storage"
	"github.com/gorilla/mux"
	"net/http"
)

type Lease struct {
}

func (l Lease) Router(router *mux.Router) {
	router.HandleFunc("/api/lease", l.List).Methods("GET")
	router.HandleFunc("/api/lease/{id}", l.List).Methods("GET")
}

func (l Lease) List(w http.ResponseWriter, r *http.Request) {
	nets := make([]schema.Lease, 0, 1024)
	for u := range storage.Network.ListLease() {
		if u == nil {
			break
		}
		nets = append(nets, *u)
	}
	ResponseJson(w, nets)
}
