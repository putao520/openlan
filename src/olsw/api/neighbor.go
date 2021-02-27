package api

import (
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/olsw/store"
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Neighbor struct {
}

func (h Neighbor) Router(router *mux.Router) {
	router.HandleFunc("/api/neighbor", h.List).Methods("GET")
}

func (h Neighbor) List(w http.ResponseWriter, r *http.Request) {
	neighbors := make([]schema.Neighbor, 0, 1024)
	for n := range store.Neighbor.List() {
		if n == nil {
			break
		}
		neighbors = append(neighbors, models.NewNeighborSchema(n))
	}
	ResponseJson(w, neighbors)
}
