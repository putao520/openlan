package api

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

func ResponseJson(w http.ResponseWriter, v interface{}) {
	str, err := json.MarshalIndent(v, "", "    ")
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func ResponseYaml(w http.ResponseWriter, v interface{}) {
	str, err := yaml.Marshal(v)
	if err == nil {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetData(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
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
