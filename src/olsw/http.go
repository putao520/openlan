package olsw

import (
	"context"
	"fmt"
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/models"
	"github.com/danieldin95/openlan-go/src/olsw/api"
	"github.com/danieldin95/openlan-go/src/olsw/storage"
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"
)

type Http struct {
	switcher   api.Switcher
	listen     string
	adminToken string
	adminFile  string
	server     *http.Server
	crtFile    string
	keyFile    string
	pubDir     string
	router     *mux.Router
}

func NewHttp(switcher api.Switcher) (h *Http) {
	c := config.Manager.Switch
	h = &Http{
		switcher:  switcher,
		listen:    c.Http.Listen,
		adminFile: c.TokenFile,
		crtFile:   c.Cert.CrtFile,
		keyFile:   c.Cert.KeyFile,
		pubDir:    c.Http.Public,
	}

	return
}

func (h *Http) Initialize() {
	r := h.Router()
	if h.server == nil {
		h.server = &http.Server{
			Addr:         h.listen,
			Handler:      r,
			ReadTimeout:  5 * time.Minute,
			WriteTimeout: 10 * time.Minute,
		}
	}
	if h.adminToken == "" {
		_ = h.LoadToken()
	}
	if h.adminToken == "" {
		h.adminToken = libol.GenRandom(32)
	}

	_ = h.SaveToken()
	h.LoadRouter()
}

func (h *Http) PProf(r *mux.Router) {
	if r != nil {
		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
}

func (h *Http) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.IsAuth(w, r) {
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic")
			http.Error(w, "Authorization Required", http.StatusUnauthorized)
		}
	})
}

func (h *Http) Router() *mux.Router {
	if h.router == nil {
		h.router = mux.NewRouter()
		h.router.Use(h.Middleware)
	}

	return h.router
}

func (h *Http) SaveToken() error {
	libol.Info("Http.SaveToken: AdminToken: %s", h.adminToken)
	f, err := os.OpenFile(h.adminFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	defer f.Close()
	if err != nil {
		libol.Error("Http.SaveToken: %s", err)
		return err
	}
	if _, err := f.Write([]byte(h.adminToken)); err != nil {
		libol.Error("Http.SaveToken: %s", err)
		return err
	}
	return nil
}

func (h *Http) LoadRouter() {
	router := h.Router()

	router.HandleFunc("/", h.IndexHtml)
	router.HandleFunc("/favicon.ico", h.PubFile)

	h.PProf(router)
	router.HandleFunc("/api/index", h.GetIndex).Methods("GET")
	router.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		format := api.GetQueryOne(r, "format")
		if format == "yaml" {
			api.ResponseYaml(w, h.switcher.Config())
		} else {
			api.ResponseJson(w, h.switcher.Config())
		}
	})
	api.Link{Switcher: h.switcher}.Router(router)
	api.User{}.Router(router)
	api.Neighbor{}.Router(router)
	api.Point{}.Router(router)
	api.Network{}.Router(router)
	api.OnLine{}.Router(router)
	api.Ctrl{Switcher: h.switcher}.Router(router)
	api.Lease{}.Router(router)
	api.Server{Switcher: h.switcher}.Router(router)
	api.Device{}.Router(router)
	api.OvClient{}.Router(router)
}

func (h *Http) LoadToken() error {
	if _, err := os.Stat(h.adminFile); os.IsNotExist(err) {
		libol.Info("Http.LoadToken: file:%s does not exist", h.adminFile)
		return nil
	}
	contents, err := ioutil.ReadFile(h.adminFile)
	if err != nil {
		libol.Error("Http.LoadToken: file:%s %s", h.adminFile, err)
		return err

	}
	h.adminToken = strings.TrimSpace(string(contents))
	return nil
}

func (h *Http) Start() {
	h.Initialize()

	libol.Info("Http.Start %s", h.listen)
	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Done(func() error {
		if h.keyFile == "" || h.crtFile == "" {
			if err := h.server.ListenAndServe(); err != nil {
				libol.Error("Http.Start on %s: %s", h.listen, err)
				return err
			}
		} else {
			if err := h.server.ListenAndServeTLS(h.crtFile, h.keyFile); err != nil {
				libol.Error("Http.Start on %s: %s", h.listen, err)
				return err
			}
		}
		return nil
	})
}

func (h *Http) Shutdown() {
	libol.Info("Http.Shutdown %s", h.listen)
	if err := h.server.Shutdown(context.Background()); err != nil {
		// Error from closing listeners, or context timeout:
		libol.Error("Http.Shutdown: %v", err)
	}
}

func (h *Http) IsAuth(w http.ResponseWriter, r *http.Request) bool {
	token, pass, ok := r.BasicAuth()
	libol.Debug("Http.IsAuth token: %s, pass: %s", token, pass)

	if path.Ext(r.URL.Path) == ".ico" {
		return true
	} else if len(r.URL.Path) > 4 {
		if !ok || token != h.adminToken {
			return false
		}
	}
	return true
}

func (h *Http) getFile(name string) string {
	return fmt.Sprintf("%s%s", h.pubDir, name)
}

func (h *Http) PubFile(w http.ResponseWriter, r *http.Request) {
	realpath := h.getFile(r.URL.Path)
	contents, err := ioutil.ReadFile(realpath)
	if err != nil {
		_, _ = fmt.Fprintf(w, "404")
		return
	}
	_, _ = fmt.Fprintf(w, "%s\n", contents)
}

func (h *Http) getIndex(body *schema.Index) *schema.Index {
	body.Version = schema.NewVersionSchema()
	body.Worker = api.NewWorkerSchema(h.switcher)

	// display accessed point.
	for p := range storage.Point.List() {
		if p == nil {
			break
		}
		body.Points = append(body.Points, models.NewPointSchema(p))
	}
	sort.SliceStable(body.Points, func(i, j int) bool {
		ii := body.Points[i]
		jj := body.Points[j]
		return ii.Network+ii.Remote > jj.Network+jj.Remote
	})
	// display neighbor.
	for n := range storage.Neighbor.List() {
		if n == nil {
			break
		}
		body.Neighbors = append(body.Neighbors, models.NewNeighborSchema(n))
	}
	sort.SliceStable(body.Neighbors, func(i, j int) bool {
		return body.Neighbors[i].IpAddr > body.Neighbors[j].IpAddr
	})
	// display links.
	for p := range storage.Link.List() {
		if p == nil {
			break
		}
		body.Links = append(body.Links, models.NewLinkSchema(p))
	}
	sort.SliceStable(body.Links, func(i, j int) bool {
		ii := body.Links[i]
		jj := body.Links[j]
		return ii.Network+ii.Server > jj.Network+jj.Server
	})
	// display online flow.
	for l := range storage.Online.List() {
		if l == nil {
			break
		}
		body.OnLines = append(body.OnLines, models.NewOnLineSchema(l))
	}
	sort.SliceStable(body.OnLines, func(i, j int) bool {
		return body.OnLines[i].HitTime < body.OnLines[j].HitTime
	})
	// display OpenVPN Clients.
	for n := range storage.Network.List() {
		if n == nil {
			break
		}
		for c := range storage.OvClient.List(n.Name) {
			if c == nil {
				break
			}
			body.OvClients = append(body.OvClients, *c)
		}
		sort.SliceStable(body.OvClients, func(i, j int) bool {
			return body.OvClients[i].Name < body.OvClients[j].Name
		})
	}
	return body
}

func (h *Http) ParseFiles(w http.ResponseWriter, name string, data interface{}) error {
	file := path.Base(name)
	tmpl, err := template.New(file).Funcs(template.FuncMap{
		"prettyTime":  libol.PrettyTime,
		"prettyBytes": libol.PrettyBytes,
		"getIpAddr":   libol.GetIPAddr,
	}).ParseFiles(name)
	if err != nil {
		_, _ = fmt.Fprintf(w, "template.ParseFiles %s", err)
		return err
	}
	if err := tmpl.Execute(w, data); err != nil {
		_, _ = fmt.Fprintf(w, "template.ParseFiles %s", err)
		return err
	}
	return nil
}

func (h *Http) IndexHtml(w http.ResponseWriter, r *http.Request) {
	body := schema.Index{
		Points:    make([]schema.Point, 0, 128),
		Links:     make([]schema.Link, 0, 128),
		Neighbors: make([]schema.Neighbor, 0, 128),
		OnLines:   make([]schema.OnLine, 0, 128),
		OvClients: make([]schema.OvClient, 0, 128),
	}
	h.getIndex(&body)
	file := h.getFile("/index.html")
	if err := h.ParseFiles(w, file, &body); err != nil {
		libol.Error("Http.Index %s", err)
	}
}

func (h *Http) GetIndex(w http.ResponseWriter, r *http.Request) {
	body := schema.Index{
		Points:    make([]schema.Point, 0, 128),
		Links:     make([]schema.Link, 0, 128),
		Neighbors: make([]schema.Neighbor, 0, 128),
		OnLines:   make([]schema.OnLine, 0, 128),
		Network:   make([]schema.Network, 0, 128),
		OvClients: make([]schema.OvClient, 0, 128),
	}
	h.getIndex(&body)
	api.ResponseJson(w, body)
}
