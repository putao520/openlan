package apiv1

import (
	"github.com/danieldin95/openlan/pkg/olctl/ctrlc"
	"github.com/danieldin95/openlan/pkg/olctl/http/api"
	"github.com/danieldin95/openlan/pkg/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Point struct {
	api.Api
}

func (p Point) Router(router *mux.Router) {
	router.HandleFunc("/api/v1/point", p.GET).Methods("GET")
	router.HandleFunc("/api/v1/point/{id}", p.GET).Methods("GET")
}

func (p Point) GET(w http.ResponseWriter, r *http.Request) {
	id, _ := api.GetArg(r, "id")
	ps := make([]schema.Point, 0, 32)
	if id == "" {
		ctrlc.Storager.Point.Iter(func(k string, v interface{}) {
			if t, ok := v.(*schema.Point); ok {
				ps = append(ps, *t)
			}
		})
	} else {
		v := ctrlc.Storager.Point.Get(id)
		if v != nil {
			if t, ok := v.(*schema.Point); ok {
				ps = append(ps, *t)
			}
		}
	}
	api.ResponseJson(w, ps)
}
