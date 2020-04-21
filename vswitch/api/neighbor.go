package api

import (
	"github.com/danieldin95/openlan-go/vswitch/schema"
	"github.com/danieldin95/openlan-go/vswitch/service"
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
	for n := range service.Neighbor.List() {
		if n == nil {
			break
		}
		neighbors = append(neighbors, schema.NewNeighbor(n))
	}
	ResponseJson(w, neighbors)
}
