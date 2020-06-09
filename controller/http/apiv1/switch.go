package apiv1

import (
	"github.com/danieldin95/openlan-go/controller/ctrlc"
	"github.com/danieldin95/openlan-go/controller/http/api"
	"github.com/danieldin95/openlan-go/switch/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Switch struct {
	api.Api
}

func (p Switch) Router(router *mux.Router) {
	router.HandleFunc("/api/v1/switch", p.GET).Methods("GET")
	router.HandleFunc("/api/v1/switch/{id}", p.GET).Methods("GET")
}

func (p Switch) GET(w http.ResponseWriter, r *http.Request) {
	id, _ := api.GetArg(r, "id")
	ss := make([]schema.Switch, 0, 32)
	if id == "" {
		ctrlc.Storager.Switch.Iter(func(k string, v interface{}) {
			if s, ok := v.(*schema.Switch); ok {
				ss = append(ss, *s)
			}
		})
	} else {
		v := ctrlc.Storager.Switch.Get(id)
		if v != nil {
			if s, ok := v.(*schema.Switch); ok {
				ss = append(ss, *s)
			}
		}
	}
	api.ResponseJson(w, ss)
}
