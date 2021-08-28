package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Apier interface {
	Router(router *mux.Router)
	GET(w http.ResponseWriter, r *http.Request)
	POST(w http.ResponseWriter, r *http.Request)
	DELETE(w http.ResponseWriter, r *http.Request)
	PUT(w http.ResponseWriter, r *http.Request)
}

type Api struct {
}

func (a Api) Router(router *mux.Router) {
	panic("implement me")
}

func (a Api) GET(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "implement me")
}

func (a Api) POST(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "implement me")
}

func (a Api) DELETE(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "implement me")
}

func (a Api) PUT(w http.ResponseWriter, r *http.Request) {
	ResponseMsg(w, 0, "implement me")
}
