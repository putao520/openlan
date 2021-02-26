package api

import (
	"github.com/danieldin95/openlan-go/src/olsw/storage"
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type OvClient struct {
}

func (h OvClient) Router(router *mux.Router) {
	router.HandleFunc("/api/ovclient", h.List).Methods("GET")
	router.HandleFunc("/api/ovclient/{id}", h.List).Methods("GET")
}

func (h OvClient) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]

	clients := make([]schema.OvClient, 0, 1024)
	if name == "" {
		for n := range storage.Network.List() {
			if n == nil {
				break
			}
			for client := range storage.OvClient.List(n.Name) {
				if client == nil {
					break
				}
				clients = append(clients, *client)
			}
		}
	} else {
		for client := range storage.OvClient.List(name) {
			if client == nil {
				break
			}
			clients = append(clients, *client)
		}
	}
	ResponseJson(w, clients)
}
