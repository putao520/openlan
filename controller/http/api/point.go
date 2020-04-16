package api

import (
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/controller/schema"
	"github.com/danieldin95/openlan-go/controller/storage"
	"github.com/gorilla/mux"
	"net/http"
)

type Point struct {
}

func (p Point) Router(router *mux.Router) {
	router.HandleFunc("/api/point", p.GET).Methods("GET")
	router.HandleFunc("/api/point/{id}", p.GET).Methods("GET")
}

func (p Point) GET(w http.ResponseWriter, r *http.Request) {
	id, _ := GetArg(r, "id")
	ps := make([]schema.Point, 0, 32)
	if id == "" {
		for vs := range storage.Storager.VSwitch.List() {
			if vs == nil {
				break
			}
			if vc, ok := vs.Ctl.(*ctl.VSwitch); ok {
				for h := range vc.ListPoint() {
					if h == nil {
						break
					}
					h.Switch = vs.Name
					ps = append(ps, *h)
				}
			}
		}
	} else {
		vs, ok := storage.Storager.VSwitch.Get(id)
		if !ok {
			http.Error(w, "switch not found", http.StatusNotFound)
			return
		}
		vc, ok := vs.Ctl.(*ctl.VSwitch)
		if !ok {
			http.Error(w, "ctl not found", http.StatusNotFound)
			return
		}
		for h := range vc.ListPoint() {
			if h == nil {
				break
			}
			h.Switch = vs.Name
			ps = append(ps, *h)
		}
	}
	ResponseJson(w, ps)
}

func (p Point) POST(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "")
}

func (p Point) PUT(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "")
}

func (p Point) DELETE(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "")
}
