package api

import (
	"github.com/danieldin95/openlan/src/libol"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type Message struct {
	Api
}

func (m Message) Router(router *mux.Router) {
	router.HandleFunc("/api/message", m.GET).Methods("GET")
}

func (m Message) GET(w http.ResponseWriter, r *http.Request) {
	size := GetQueryOne(r, "size")
	us := make([]libol.Message, 0, 32)
	for h := range libol.Logger.List() {
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
