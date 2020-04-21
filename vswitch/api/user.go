package api

import (
	"encoding/json"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/vswitch/schema"
	"github.com/danieldin95/openlan-go/vswitch/service"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

type User struct {
}

func (h User) Router(router *mux.Router) {
	router.HandleFunc("/api/user", h.List).Methods("GET")
	router.HandleFunc("/api/user/{id}", h.Get).Methods("GET")
	router.HandleFunc("/api/user/{id}", h.Add).Methods("POST")
	router.HandleFunc("/api/user/{id}", h.Del).Methods("DELETE")
}

func (h User) List(w http.ResponseWriter, r *http.Request) {
	users := make([]schema.User, 0, 1024)
	for u := range service.User.List() {
		if u == nil {
			break
		}
		users = append(users, schema.NewUser(u))
	}
	ResponseJson(w, users)
}

func (h User) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := service.User.Get(vars["id"])
	if user != nil {
		ResponseJson(w, schema.NewUser(user))
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h User) Add(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := &schema.User{}
	if err := json.Unmarshal([]byte(body), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	service.User.Add(user.ToModel())
	ResponseMsg(w, 0, "")
}

func (h User) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("DelUser %s", vars["id"])

	service.User.Del(vars["id"])
	ResponseMsg(w, 0, "")
}
