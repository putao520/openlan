package api

import (
	"github.com/danieldin95/openlan/pkg/models"
	"github.com/danieldin95/openlan/pkg/olsw/cache"
	"github.com/danieldin95/openlan/pkg/schema"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type Network struct {
}

func (h Network) Router(router *mux.Router) {
	router.HandleFunc("/api/network", h.List).Methods("GET")
	router.HandleFunc("/api/network/{id}", h.Get).Methods("GET")
	router.HandleFunc("/get/network/{id}/{ie}.ovpn", h.Profile).Methods("GET")
}

func (h Network) List(w http.ResponseWriter, r *http.Request) {
	nets := make([]schema.Network, 0, 1024)
	for u := range cache.Network.List() {
		if u == nil {
			break
		}
		nets = append(nets, models.NewNetworkSchema(u))
	}
	ResponseJson(w, nets)
}

func (h Network) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	net := cache.Network.Get(vars["id"])
	if net != nil {
		ResponseJson(w, models.NewNetworkSchema(net))
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h Network) Profile(w http.ResponseWriter, r *http.Request) {
	server := strings.SplitN(r.Host, ":", 2)[0]
	vars := mux.Vars(r)
	data, err := cache.VPNClient.GetClientProfile(vars["id"], vars["ie"], server)
	if err == nil {
		_, _ = w.Write([]byte(data))
	} else {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
