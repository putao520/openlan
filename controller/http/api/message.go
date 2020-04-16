package api

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type Message struct {
}

func (m Message) Router(router *mux.Router) {
	router.HandleFunc("/api/message", m.GET).Methods("GET")
}

func (m Message) GET(w http.ResponseWriter, r *http.Request) {
	size := GetQueryOne(r, "size")
	us := make([]libstar.Message, 0, 32)
	for h := range libstar.Log.List() {
		if h == nil {
			break
		}
		us = append(us, *h)
	}

	total := len(us)
	if s, err := strconv.Atoi(size); err == nil {
		if total > s {
			us = us[:s]
		} else {
			us = us[:total]
		}
	}
	ResponseJson(w, us)
}
