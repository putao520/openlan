package api

import (
	"encoding/json"
	"github.com/danieldin95/openlan/pkg/config"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/models"
	"github.com/danieldin95/openlan/pkg/olsw/cache"
	"github.com/danieldin95/openlan/pkg/schema"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

type Link struct {
	Switcher Switcher
}

func (h Link) Router(router *mux.Router) {
	router.HandleFunc("/api/link", h.List).Methods("GET")
	router.HandleFunc("/api/link/{id}", h.Get).Methods("GET")
	router.HandleFunc("/api/link/{id}", h.Add).Methods("POST")
	router.HandleFunc("/api/link/{id}", h.Del).Methods("DELETE")
}

func (h Link) List(w http.ResponseWriter, r *http.Request) {
	links := make([]schema.Link, 0, 1024)
	for l := range cache.Link.List() {
		if l == nil {
			break
		}
		links = append(links, models.NewLinkSchema(l))
	}
	ResponseJson(w, links)
}

func (h Link) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("GetPoint %s", vars["id"])

	link := cache.Link.Get(vars["id"])
	if link != nil {
		ResponseJson(w, models.NewLinkSchema(link))
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h Link) Add(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c := &config.Point{}
	if err := json.Unmarshal(body, c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Default()
	h.Switcher.AddLink(c.Network, c)
	ResponseMsg(w, 0, "")
}

func (h Link) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("DelLink %s", vars["id"])
	h.Switcher.DelLink("", vars["id"])
	ResponseMsg(w, 0, "")
}
