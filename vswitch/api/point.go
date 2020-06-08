package api

import (
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/schema"
	"github.com/danieldin95/openlan-go/vswitch/storage"
	"github.com/gorilla/mux"
	"net/http"
)

type Point struct {
}

func (h Point) Router(router *mux.Router) {
	router.HandleFunc("/api/point", h.List).Methods("GET")
	router.HandleFunc("/api/point/{id}", h.Get).Methods("GET")
}

func (h Point) List(w http.ResponseWriter, r *http.Request) {
	points := make([]schema.Point, 0, 1024)
	for u := range storage.Point.List() {
		if u == nil {
			break
		}
		points = append(points, models.NewPointSchema(u))
	}
	ResponseJson(w, points)
}

func (h Point) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	point := storage.Point.Get(vars["id"])
	if point != nil {
		ResponseJson(w, models.NewPointSchema(point))
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}
