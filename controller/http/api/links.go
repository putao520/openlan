package api

import (
	"github.com/gorilla/mux"
)

type Link struct {
	Api
}

func (l Link) Router(router *mux.Router) {
	router.HandleFunc("/api/link/{id}", l.GET).Methods("GET")
}
