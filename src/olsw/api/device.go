package api

import (
	"github.com/danieldin95/openlan-go/src/network"
	"github.com/danieldin95/openlan-go/src/olsw/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Device struct {
}

func (h Device) Router(router *mux.Router) {
	router.HandleFunc("/api/device", h.List).Methods("GET")
	router.HandleFunc("/api/device/{id}", h.Get).Methods("GET")
}

func (h Device) List(w http.ResponseWriter, r *http.Request) {
	dev := make([]schema.Device, 0, 1024)
	for t := range network.Tapers.List() {
		if t == nil {
			break
		}
		st := schema.Device{
			Name: t.Name(),
			Mtu:  t.Mtu(),
		}
		if t.IsTun() {
			st.Type = "tun"
		} else if t.IsTap() {
			st.Type = "tap"
		}
		dev = append(dev, st)
	}
	ResponseJson(w, dev)
}

func (h Device) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dev := network.Tapers.Get(vars["id"])
	if dev != nil {
		ResponseJson(w, schema.Device{
			Name: dev.Name(),
			Mtu:  dev.Mtu(),
		})
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}
