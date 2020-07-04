package api

import (
	"github.com/danieldin95/openlan-go/src/controller/ctrlc"
	"github.com/danieldin95/openlan-go/src/switch/schema"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
)

type Switch struct {
	Api
}

func (z Switch) Router(router *mux.Router) {
	router.HandleFunc("/api/switch", z.GET).Methods("GET")
}

func (z Switch) GET(w http.ResponseWriter, r *http.Request) {
	ss := make([]schema.Switch, 0, 32)
	ctrlc.Storager.Switch.Iter(func(k string, v interface{}) {
		if s, ok := v.(*schema.Switch); ok {
			ss = append(ss, *s)
		}
	})
	sort.SliceStable(ss, func(i, j int) bool {
		return ss[i].Alias < ss[j].Alias
	})
	ResponseJson(w, ss)
}
