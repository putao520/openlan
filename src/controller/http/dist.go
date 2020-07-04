package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type Dist struct {
	Dir string
}

func (s Dist) Router(router *mux.Router) {
	// dist
	dir := http.Dir(s.Dir)
	distFile := http.StripPrefix("/dist/", http.FileServer(dir))
	router.PathPrefix("/dist/").Handler(distFile)

	// js
	jsDir := http.Dir(s.Dir + "/js")
	jsFile := http.StripPrefix("/js/", http.FileServer(jsDir))
	router.PathPrefix("/js/").Handler(jsFile)

	// css
	cssDir := http.Dir(s.Dir + "/css")
	cssFile := http.StripPrefix("/css/", http.FileServer(cssDir))
	router.PathPrefix("/css/").Handler(cssFile)

	// fonts
	fontsDir := http.Dir(s.Dir + "/fonts")
	fontsFile := http.StripPrefix("/fonts/", http.FileServer(fontsDir))
	router.PathPrefix("/fonts/").Handler(fontsFile)

	// root
	router.HandleFunc("/", s.Ui)
	router.HandleFunc("/dist", s.Ui)
}

func (s Dist) GetFile(name string) string {
	return fmt.Sprintf("%s/%s", s.Dir, name)
}

func (s Dist) Ui(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dist/", http.StatusTemporaryRedirect)
}
