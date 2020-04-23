package apiv1

import (
	"github.com/danieldin95/openlan-go/controller/ctl"
	"github.com/danieldin95/openlan-go/controller/http/api"
	"github.com/danieldin95/openlan-go/controller/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Point struct {
}

func (p Point) Router(router *mux.Router) {
	router.HandleFunc("/api/v1/point", p.GET).Methods("GET")
	router.HandleFunc("/api/v1/point/{id}", p.GET).Methods("GET")
}

func (p Point) GET(w http.ResponseWriter, r *http.Request) {
	id, _ := api.GetArg(r, "id")
	ps := make([]schema.Point, 0, 32)
	if id == "" {
		ctl.Storager.Point.Iter(func(k string, v interface{}) {
			if p, ok := v.(*schema.Point); ok {
				ps = append(ps, *p)
			}
		})
	} else {
		v := ctl.Storager.Point.Get(id)
		if v != nil {
			if p, ok := v.(*schema.Point); ok {
				ps = append(ps, *p)
			}
		}
	}
	api.ResponseJson(w, ps)
}

func (p Point) POST(w http.ResponseWriter, r *http.Request) {
	api.ResponseMsg(w, 0, "")
}

func (p Point) PUT(w http.ResponseWriter, r *http.Request) {
	api.ResponseMsg(w, 0, "")
}

func (p Point) DELETE(w http.ResponseWriter, r *http.Request) {
	api.ResponseMsg(w, 0, "")
}
