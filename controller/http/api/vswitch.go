package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

type VSwitch struct {
	Api
}

func (z VSwitch) Router(router *mux.Router) {
	router.HandleFunc("/api/vswitch", z.GET).Methods("GET")
}

func (z VSwitch) GET(w http.ResponseWriter, r *http.Request) {
	//vs := make([]schema.VSwitch, 0, 32)
	//for h := range storage.Storager.VSwitch.List() {
	//	if h == nil {
	//		break
	//	}
	//	if vc, ok := h.Ctl.(*ctl.VSwitch); ok {
	//		h.State = vc.State
	//	}
	//	vs = append(vs, *h)
	//}
	//sort.SliceStable(vs, func(i, j int) bool {
	//	return vs[i].Name < vs[j].Name
	//})
	//ResponseJson(w, vs)
}
