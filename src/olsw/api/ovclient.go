package api

import (
	"github.com/danieldin95/openlan-go/src/olsw/schema"
	"github.com/danieldin95/openlan-go/src/olsw/storage"
	"github.com/gorilla/mux"
	"net/http"
)

type OvClient struct {
}

func (h OvClient) Router(router *mux.Router) {
	router.HandleFunc("/api/ovclient/{id}", h.List).Methods("GET")
}

func (h OvClient) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	clients := make([]schema.OvClient, 0, 1024)
	for client := range storage.OvClient.List(name) {
		if client == nil {
			break
		}
		clients = append(clients, *client)
	}
	ResponseJson(w, clients)
}
