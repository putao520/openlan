package api

import (
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/controller/schema"
	"github.com/danieldin95/openlan-go/controller/storage"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
)

type VSwitch struct {
}

func (z VSwitch) Router(router *mux.Router) {
	router.HandleFunc("/api/vswitch", z.GET).Methods("GET")
}

func (z VSwitch) GET(w http.ResponseWriter, r *http.Request) {
	vs := make([]schema.VSwitch, 0, 32)
	for h := range storage.Storager.VSwitch.List() {
		if h == nil {
			break
		}
		if vc, ok := h.Ctl.(*ctl.VSwitch); ok {
			h.State = vc.State
		}
		vs = append(vs, *h)
	}
	sort.SliceStable(vs, func(i, j int) bool {
		return vs[i].Name < vs[j].Name
	})
	ResponseJson(w, vs)
}

func (z VSwitch) POST(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "")
}

func (z VSwitch) PUT(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "")
}

func (z VSwitch) DELETE(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "")
}
