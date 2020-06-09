package http

import (
	"context"
	"github.com/danieldin95/openlan-go/controller/http/api"
	"github.com/danieldin95/openlan-go/controller/http/apiv1"
	"github.com/danieldin95/openlan-go/controller/http/olan"
	"github.com/danieldin95/openlan-go/controller/storage"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strings"
)

type Server struct {
	listen  string
	server  *http.Server
	crtFile string
	keyFile string
	pubDir  string
	router  *mux.Router
}

func NewServer(listen, staticDir string) (h *Server) {
	h = &Server{
		listen: listen,
		pubDir: staticDir,
	}
	return
}

func (h *Server) Router() *mux.Router {
	if h.router != nil {
		return h.router
	}
	h.router = mux.NewRouter()
	return h.router
}

func (h *Server) LoadRouter() {
	router := h.Router()
	router.Use(h.Middleware)

	// API legacy
	api.Switch{}.Router(router)
	api.User{}.Router(router)
	api.Point{}.Router(router)
	api.Graph{}.Router(router)
	api.Message{}.Router(router)
	// API V1
	apiv1.Point{}.Router(router)
	apiv1.Switch{}.Router(router)
	apiv1.Neighbor{}.Router(router)
	// Static files
	Dist{h.pubDir}.Router(router)
	// OpenLAN message
	olan.Ctrl{}.Router(router)
}

func (h *Server) SetCert(keyFile, crtFile string) {
	h.crtFile = crtFile
	h.keyFile = keyFile
}

func (h *Server) Initialize() {
	r := h.Router()
	if h.server == nil {
		h.server = &http.Server{
			Addr:    h.listen,
			Handler: r,
		}
	}
	h.LoadRouter()
}

func (h *Server) IsAuth(w http.ResponseWriter, r *http.Request) bool {
	name, pass, _ := api.GetAuth(r)
	libol.Print("Server.IsAuth %s:%s", name, pass)

	user, ok := storage.Storager.Users.Get(name)
	if !ok || user.Password != pass {
		return false
	}
	return true
}

func (h *Server) LogRequest(r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api") || r.Method == "GET" {
		libol.Debug("Server.Middleware %s %s %s", r.RemoteAddr, r.Method, r.URL)
		return
	}
	path := r.URL.Path
	if q, _ := url.QueryUnescape(r.URL.RawQuery); q != "" {
		path += "?" + q
	}
	libol.Info("Server.Middleware %s %s %s", r.RemoteAddr, r.Method, path)
}

func (h *Server) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.LogRequest(r)
		if h.IsAuth(w, r) {
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic")
			http.Error(w, "Authorization Required.", http.StatusUnauthorized)
		}
	})
}

func (h *Server) Start() error {
	h.Initialize()
	if h.keyFile == "" || h.crtFile == "" {
		libol.Info("Server.Start http://%s", h.listen)
		if err := h.server.ListenAndServe(); err != nil {
			libol.Error("Server.Start on %s: %s", h.listen, err)
			return err
		}
	} else {
		libol.Info("Server.Start https://%s", h.listen)
		if err := h.server.ListenAndServeTLS(h.crtFile, h.keyFile); err != nil {
			libol.Error("Server.Start on %s: %s", h.listen, err)
			return err
		}
	}
	return nil
}

func (h *Server) Shutdown() {
	libol.Info("Server.Shutdown %s", h.listen)
	if err := h.server.Shutdown(context.Background()); err != nil {
		libol.Error("Server.Shutdown %v", err)
	}
}
