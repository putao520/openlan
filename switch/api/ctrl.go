package api

import (
	"github.com/danieldin95/openlan-go/switch/ctrls"
	"github.com/danieldin95/openlan-go/switch/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Ctrl struct {
	Switcher Switcher
}

func (h Ctrl) Router(router *mux.Router) {
	router.HandleFunc("/api/ctrl", h.Get).Methods("GET")
	router.HandleFunc("/api/ctrl", h.Add).Methods("POST")
	router.HandleFunc("/api/ctrl", h.Del).Methods("DELETE")
}

func (h Ctrl) Get(w http.ResponseWriter, r *http.Request) {
	ResponseJson(w, ctrls.Ctrl)
}

func (h Ctrl) Add(w http.ResponseWriter, r *http.Request) {
	conf := &schema.Ctrl{}
	if err := GetData(r, conf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctrls.Ctrl.Stop()
	ctrls.Ctrl = &ctrls.CtrlC{
		Url:      conf.Url,
		Name:     h.Switcher.Alias(),
		Password: conf.Token,
	}
	err := ctrls.Ctrl.Open()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctrls.Ctrl.Start()
}

func (h Ctrl) Del(w http.ResponseWriter, r *http.Request) {
	ctrls.Ctrl.Stop()
}
