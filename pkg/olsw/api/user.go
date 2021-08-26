package api

import (
	"encoding/json"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/models"
	"github.com/danieldin95/openlan/pkg/olsw/store"
	"github.com/danieldin95/openlan/pkg/schema"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"sort"
)

type User struct {
}

func (h User) Router(router *mux.Router) {
	router.HandleFunc("/api/user", h.List).Methods("GET")
	router.HandleFunc("/api/user", h.Add).Methods("POST")
	router.HandleFunc("/api/user/{id}", h.Get).Methods("GET")
	router.HandleFunc("/api/user/{id}", h.Add).Methods("POST")
	router.HandleFunc("/api/user/{id}", h.Del).Methods("DELETE")
	router.HandleFunc("/api/user/{id}/check", h.Check).Methods("POST")
}

func (h User) List(w http.ResponseWriter, r *http.Request) {
	users := make([]schema.User, 0, 1024)
	for u := range store.User.List() {
		if u == nil {
			break
		}
		users = append(users, models.NewUserSchema(u))
	}
	sort.SliceStable(users, func(i, j int) bool {
		return users[i].Network+users[i].Name > users[j].Network+users[j].Name
	})
	ResponseJson(w, users)
}

func (h User) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := store.User.Get(vars["id"])
	if user != nil {
		ResponseJson(w, models.NewUserSchema(user))
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
	if err := json.Unmarshal(body, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	store.User.Add(models.SchemaToUserModel(user))
	if err := store.User.Save(); err != nil {
		libol.Warn("AddUser %s", err)
	}
	ResponseMsg(w, 0, "")
}

func (h User) Del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("DelUser %s", vars["id"])

	store.User.Del(vars["id"])
	if err := store.User.Save(); err != nil {
		libol.Warn("DelUser %s", err)
	}
	ResponseMsg(w, 0, "")
}

func (h User) Check(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user := &schema.User{}
	if err := json.Unmarshal(body, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	model := models.SchemaToUserModel(user)
	if obj := store.User.Check(model); obj != nil {
		ResponseJson(w, models.NewUserSchema(obj))
	} else {
		http.Error(w, "invalid user", http.StatusUnauthorized)
		return
	}
}
