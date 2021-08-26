package api

import (
	"encoding/base64"
	"encoding/json"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/olctl/storage"
	"github.com/danieldin95/openlan/pkg/schema"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strings"
)

func GetArg(r *http.Request, name string) (string, bool) {
	vars := mux.Vars(r)
	value, ok := vars[name]
	return value, ok
}

func GetData(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(body), v); err != nil {
		return err
	}
	return nil
}

func GetQueryOne(req *http.Request, name string) string {
	query := req.URL.Query()
	if values, ok := query[name]; ok {
		return values[0]
	}
	return ""
}

func ResponseJson(w http.ResponseWriter, v interface{}) {
	str, err := json.Marshal(v)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ResponseXML(w http.ResponseWriter, v string) {
	w.Header().Set("Content-Type", "application/xml")
	_, _ = w.Write([]byte(v))
}

func ResponseMsg(w http.ResponseWriter, code int, message string) {
	ret := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    code,
		Message: message,
	}
	ResponseJson(w, ret)
}

func GetAuth(req *http.Request) (name, pass string, ok bool) {
	if t, err := req.Cookie("serve-token-jhx"); err == nil {
		name, pass, ok = ParseBasicAuth(t.Value)
	} else {
		name, pass, ok = req.BasicAuth()
	}
	return name, pass, ok
}

func GetUser(req *http.Request) (schema.User, bool) {
	name, _, _ := GetAuth(req)
	libol.Debug("GetUser %s", name)
	return storage.Storager.Users.Get(name)
}

func ParseBasicAuth(auth string) (username, password string, ok bool) {
	c, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}
