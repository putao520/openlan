package api

import (
	"github.com/danieldin95/openlan-go/src/olctl/ctrlc"
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/gorilla/mux"
	"net/http"
)

type Graph struct {
	Api
}

func (g Graph) Router(router *mux.Router) {
	router.HandleFunc("/api/graph/{id}", g.GET).Methods("GET")
}

func (g Graph) GET(w http.ResponseWriter, r *http.Request) {
	//id, _ := GetArg(r, "id")
	graphs := struct {
		Categories []schema.Category   `json:"categories"`
		Nodes      []*schema.GraphNode `json:"nodes"`
		Links      []*schema.GraphLink `json:"links"`
	}{
		Categories: []schema.Category{
			{Name: "virtual switch"},
			{Name: "accessed point"},
		},
		Nodes: make([]*schema.GraphNode, 0, 32),
		Links: make([]*schema.GraphLink, 0, 32),
	}

	i := 0
	nn := make(map[string]*schema.GraphNode, 32)
	ctrlc.Storager.Switch.Iter(func(k string, v interface{}) {
		s, ok := v.(*schema.Switch)
		if ok {
			node := &schema.GraphNode{
				Name:       s.Alias,
				SymbolSize: 15,
				Category:   0,
				Id:         i,
			}
			nn[s.Alias] = node
			graphs.Nodes = append(graphs.Nodes, node)
			i += 1
		}
	})
	ctrlc.Storager.Point.Iter(func(k string, v interface{}) {
		p, ok := v.(*schema.Point)
		if !ok {
			return
		}
		sn, ok := nn[p.Switch]
		if !ok {
			return
		}
		pn, ok := nn[p.Alias]
		if !ok {
			pn = &schema.GraphNode{
				Name:       p.Alias,
				SymbolSize: 10,
				Category:   1,
				Id:         i,
			}
			nn[p.Alias] = pn
			graphs.Nodes = append(graphs.Nodes, pn)
			i += 1
		}
		graphs.Links = append(graphs.Links, &schema.GraphLink{
			Source: pn.Id,
			Target: sn.Id,
		})
	})
	ResponseJson(w, graphs)
}
