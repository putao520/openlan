package api

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	Switcher Switcher
}

func (l Server) Router(router *mux.Router) {
	router.HandleFunc("/api/server", l.List).Methods("GET")
	router.HandleFunc("/api/server/{id}", l.List).Methods("GET")
}

func (l Server) List(w http.ResponseWriter, r *http.Request) {
	server := l.Switcher.Server()
	data := &struct {
		UpTime     int64           `json:"uptime"`
		Statistic  libol.ServerSts `json:"statistic"`
		Connection []interface{}
	}{
		UpTime:     l.Switcher.UpTime(),
		Statistic:  server.Sts(),
		Connection: make([]interface{}, 0, 1024),
	}
	for u := range server.ListClient() {
		if u == nil {
			break
		}
		data.Connection = append(data.Connection, &struct {
			UpTime     int64           `json:"uptime"`
			LocalAddr  string          `json:"localAddr"`
			RemoteAddr string          `json:"remoteAddr"`
			Statistic  libol.ClientSts `json:"statistic"`
		}{
			UpTime:     u.UpTime(),
			LocalAddr:  u.LocalAddr(),
			RemoteAddr: u.RemoteAddr(),
			Statistic:  u.Sts(),
		})
	}
	ResponseJson(w, data)
}
