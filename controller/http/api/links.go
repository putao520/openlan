package api

import (
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/controller/schema"
	"github.com/danieldin95/openlan-go/controller/storage"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
)

type Link struct {
}

func (l Link) Router(router *mux.Router) {
	router.HandleFunc("/api/link/{id}", l.GET).Methods("GET")
}

func (l Link) GET(w http.ResponseWriter, r *http.Request) {
	id, _ := GetArg(r, "id")
	vs, ok := storage.Storager.VSwitch.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	vc, ok := vs.Ctl.(*ctl.VSwitch)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	ls := make([]schema.Point, 0, 32)
	for h := range vc.ListLink() {
		if h == nil {
			break
		}
		ls = append(ls, *h)
	}
	sort.SliceStable(ls, func(i, j int) bool {
		return ls[i].Alias < ls[j].Alias
	})
	ResponseJson(w, ls)
}

func (l Link) POST(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "")
}

func (l Link) PUT(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "")
}

func (l Link) DELETE(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "")
}
