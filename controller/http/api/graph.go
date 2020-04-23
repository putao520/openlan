package api

import (
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
	//gs := struct {
	//	Categories []schema.Category `json:"categories"`
	//	Nodes      []*schema.Node    `json:"nodes"`
	//	Links      []*schema.Link    `json:"links"`
	//}{
	//	Categories: []schema.Category{
	//		{
	//			Name: "virtual switch",
	//		},
	//		{
	//			Name: "accessed point",
	//		},
	//	},
	//	Nodes: make([]*schema.Node, 0, 32),
	//	Links: make([]*schema.Link, 0, 32),
	//}
	//
	//i := 0
	//idx := make(map[string]*schema.Node, 32)
	//for vs := range storage.Storager.VSwitch.List() {
	//	if vs == nil {
	//		break
	//	}
	//	node := &schema.Node{
	//		Name:       vs.Name,
	//		SymbolSize: 20,
	//		Category:   0,
	//		Id:         i,
	//	}
	//	idx[vs.Name] = node
	//	gs.Nodes = append(gs.Nodes, node)
	//	i += 1
	//}
	//for vs := range storage.Storager.VSwitch.List() {
	//	if vs == nil {
	//		break
	//	}
	//	vnode, ok := idx[vs.Name]
	//	if !ok {
	//		continue
	//	}
	//	vc, ok := vs.Ctl.(*ctl.VSwitch)
	//	if !ok {
	//		continue
	//	}
	//	for p := range vc.ListPoint() {
	//		if p == nil {
	//			break
	//		}
	//		pnode, ok := idx[p.Alias]
	//		if !ok {
	//			pnode = &schema.Node{
	//				Name:       p.Alias,
	//				SymbolSize: 10,
	//				Category:   1,
	//				Id:         i,
	//			}
	//			idx[p.Alias] = pnode
	//			gs.Nodes = append(gs.Nodes, pnode)
	//			i += 1
	//		}
	//		vnode.SymbolSize += 1
	//		gs.Links = append(gs.Links, &schema.Link{
	//			Source: pnode.Id,
	//			Target: vnode.Id,
	//		})
	//	}
	//	// TODO links.
	//}
	//ResponseJson(w, gs)
}
