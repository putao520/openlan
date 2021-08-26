package api

import (
	"github.com/danieldin95/openlan/pkg/olctl/ctrlc"
	"github.com/danieldin95/openlan/pkg/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Point struct {
	Api
}

func (p Point) Router(router *mux.Router) {
	router.HandleFunc("/api/point", p.GET).Methods("GET")
	router.HandleFunc("/api/point/{id}", p.GET).Methods("GET")
}

func (p Point) GET(w http.ResponseWriter, r *http.Request) {
	id, _ := GetArg(r, "id")
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
	ResponseJson(w, ps)
}
