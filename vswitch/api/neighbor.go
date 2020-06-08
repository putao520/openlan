package api

import (
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/schema"
	"github.com/danieldin95/openlan-go/vswitch/storage"
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
	for n := range storage.Neighbor.List() {
		if n == nil {
			break
		}
		neighbors = append(neighbors, models.NewNeighborSchema(n))
	}
	ResponseJson(w, neighbors)
}
