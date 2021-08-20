package api

import (
	"github.com/danieldin95/openlan/src/models"
	"github.com/danieldin95/openlan/src/olsw/store"
	"github.com/danieldin95/openlan/src/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Esp struct {
	Switcher Switcher
}

func (l Esp) Router(router *mux.Router) {
	router.HandleFunc("/api/esp", l.List).Methods("GET")
	router.HandleFunc("/api/esp/{id}", l.List).Methods("GET")
}

func (l Esp) List(w http.ResponseWriter, r *http.Request) {
	data := make([]schema.Esp, 0, 1024)
	for e := range store.Esp.List() {
		if e == nil {
			break
		}
		item := models.NewEspSchema(e)
		data = append(data, item)
	}
	ResponseJson(w, data)
}

type EspState struct {
	Switcher Switcher
}

func (l EspState) Router(router *mux.Router) {
	router.HandleFunc("/api/state", l.List).Methods("GET")
	router.HandleFunc("/api/state/{id}", l.List).Methods("GET")
}

func (l EspState) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]
	data := make([]schema.EspState, 0, 1024)
	for e := range store.EspState.List(name) {
		if e == nil {
			break
		}
		data = append(data, models.NewEspStateSchema(e))
	}
	ResponseJson(w, data)
}

type EspPolicy struct {
	Switcher Switcher
}

func (l EspPolicy) Router(router *mux.Router) {
	router.HandleFunc("/api/policy", l.List).Methods("GET")
	router.HandleFunc("/api/policy/{id}", l.List).Methods("GET")
}

func (l EspPolicy) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["id"]
	data := make([]schema.EspPolicy, 0, 1024)
	for e := range store.EspPolicy.List(name) {
		if e == nil {
			break
		}
		data = append(data, models.NewEspPolicySchema(e))
	}
	ResponseJson(w, data)
}
