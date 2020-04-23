package api

import (
	"github.com/danieldin95/openlan-go/controller/schema"
	"github.com/danieldin95/openlan-go/controller/storage"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
)

type User struct {
	Api
}

func (u User) Router(router *mux.Router) {
	router.HandleFunc("/api/user", u.GET).Methods("GET")
}

func (u User) GET(w http.ResponseWriter, r *http.Request) {
	us := make([]schema.User, 0, 32)
	for h := range storage.Storager.Users.List() {
		if h == nil {
			break
		}
		us = append(us, *h)
	}
	sort.SliceStable(us, func(i, j int) bool {
		return us[i].Name < us[j].Name
	})
	ResponseJson(w, us)
}
