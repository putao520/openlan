package api

import (
	"github.com/danieldin95/openlan-go/src/controller/ctrlc"
	"github.com/danieldin95/openlan-go/src/controller/schema"
	schema1 "github.com/danieldin95/openlan-go/src/switch/schema"
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
		Categories []schema.Category `json:"categories"`
		Nodes      []*schema.Node    `json:"nodes"`
		Links      []*schema.Link    `json:"links"`
	}{
		Categories: []schema.Category{
			{Name: "virtual switch"},
			{Name: "accessed point"},
		},
		Nodes: make([]*schema.Node, 0, 32),
		Links: make([]*schema.Link, 0, 32),
	}

	i := 0
	nn := make(map[string]*schema.Node, 32)
	ctrlc.Storager.Switch.Iter(func(k string, v interface{}) {
		s, ok := v.(*schema1.Switch)
		if ok {
			node := &schema.Node{
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
		p, ok := v.(*schema1.Point)
		if !ok {
			return
		}
		vn, ok := nn[p.Switch]
		if !ok {
			return
		}
		pn, ok := nn[p.Alias]
		if !ok {
			pn = &schema.Node{
				Name:       p.Alias,
				SymbolSize: 10,
				Category:   1,
				Id:         i,
			}
			nn[p.Alias] = pn
			graphs.Nodes = append(graphs.Nodes, pn)
			i += 1
		}
		vn.SymbolSize += 1
		graphs.Links = append(graphs.Links, &schema.Link{
			Source: pn.Id,
			Target: vn.Id,
		})
	})
	ResponseJson(w, graphs)
}
