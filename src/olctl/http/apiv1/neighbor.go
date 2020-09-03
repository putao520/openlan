package apiv1

import (
	"github.com/danieldin95/openlan-go/src/olctl/ctrlc"
	"github.com/danieldin95/openlan-go/src/olctl/http/api"
	"github.com/danieldin95/openlan-go/src/olsw/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Neighbor struct {
	api.Api
}

func (p Neighbor) Router(router *mux.Router) {
	router.HandleFunc("/api/v1/neighbor", p.GET).Methods("GET")
	router.HandleFunc("/api/v1/neighbor/{id}", p.GET).Methods("GET")
}

func (p Neighbor) GET(w http.ResponseWriter, r *http.Request) {
	id, _ := api.GetArg(r, "id")
	ss := make([]schema.Neighbor, 0, 32)
	if id == "" {
		ctrlc.Storager.Neighbor.Iter(func(k string, v interface{}) {
			if s, ok := v.(*schema.Neighbor); ok {
				ss = append(ss, *s)
			}
		})
	} else {
		v := ctrlc.Storager.Neighbor.Get(id)
		if v != nil {
			if s, ok := v.(*schema.Neighbor); ok {
				ss = append(ss, *s)
			}
		}
	}
	api.ResponseJson(w, ss)
}
