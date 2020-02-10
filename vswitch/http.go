package vswitch

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/models"
	"github.com/danieldin95/openlan-go/vswitch/service"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"
)

type Http struct {
	worker     *Worker
	listen     string
	adminToken string
	adminFile  string
	server     *http.Server
	crtFile    string
	keyFile    string
	pubDir     string
	router     *mux.Router
}

func NewHttp(worker *Worker, c *config.VSwitch) (h *Http) {
	h = &Http{
		worker:     worker,
		listen:     c.HttpListen,
		adminToken: c.Token,
		adminFile:  c.TokenFile,
		crtFile:    c.CrtFile,
		keyFile:    c.KeyFile,
		pubDir:     c.HttpDir,
	}

	return
}

func (h *Http) Initialize() {
	r := h.Router()
	if h.server == nil {
		h.server = &http.Server{
			Addr:         h.listen,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      r,
		}
	}

	if h.adminToken == "" {
		h.LoadToken()
	}

	if h.adminToken == "" {
		h.adminToken = libol.GenToken(64)
	}

	h.SaveToken()
	h.LoadRouter()
}

func (h *Http) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.IsAuth(w, r) {
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic")
			http.Error(w, "Authorization Required.", http.StatusUnauthorized)
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
	h.Router().HandleFunc("/", h.IndexHtml)
	h.Router().HandleFunc("/favicon.ico", h.PubFile)
	h.Router().HandleFunc("/api/link", h.ListLink).Methods("GET")
	h.Router().HandleFunc("/api/link/{id}", h.GetLink).Methods("GET")
	h.Router().HandleFunc("/api/link/{id}", h.AddLink).Methods("POST")
	h.Router().HandleFunc("/api/link/{id}", h.DelLink).Methods("DELETE")
	h.Router().HandleFunc("/api/user", h.ListUser).Methods("GET")
	h.Router().HandleFunc("/api/user/{id}", h.GetUser).Methods("GET")
	h.Router().HandleFunc("/api/user/{id}", h.AddUser).Methods("POST")
	h.Router().HandleFunc("/api/user/{id}", h.DelUser).Methods("DELETE")
	h.Router().HandleFunc("/api/neighbor", h.ListNeighbor).Methods("GET")
	h.Router().HandleFunc("/api/point", h.ListPoint).Methods("GET")
	h.Router().HandleFunc("/api/point/{id}", h.GetPoint).Methods("GET")
	h.Router().HandleFunc("/api/network", h.ListNetwork).Methods("GET")
	h.Router().HandleFunc("/api/network/{id}", h.GetNetwork).Methods("GET")
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

	h.adminToken = string(contents)
	return nil
}

func (h *Http) Start() error {
	h.Initialize()

	libol.Info("Http.Start %s", h.listen)

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

	if len(r.URL.Path) < 4 || r.URL.Path[:4] != "/api" {
		return true
	}

	if !ok || token != h.adminToken {
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Authorization Required.", http.StatusUnauthorized)
		return false
	}

	return true
}

func (h *Http) ResponseJson(w http.ResponseWriter, v interface{}) {
	str, err := json.Marshal(v)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Http) ResponseMsg(w http.ResponseWriter, code int, message string) {
	ret := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    code,
		Message: message,
	}
	h.ResponseJson(w, ret)
}

func (h *Http) getFile(name string) string {
	return fmt.Sprintf("%s%s", h.pubDir, name)
}

func (h *Http) PubFile(w http.ResponseWriter, r *http.Request) {
	realpath := h.getFile(r.URL.Path)
	contents, err := ioutil.ReadFile(realpath)
	if err != nil {
		fmt.Fprintf(w, "404")
		return
	}

	fmt.Fprintf(w, "%s\n", contents)
}

func (h *Http) getIndex(body *IndexSchema) *IndexSchema {
	body.Worker.UUID = h.worker.UUID()
	body.Worker.Uptime = h.worker.UpTime()

	pointList := make([]*models.Point, 0, 128)
	for p := range service.Point.List() {
		if p == nil {
			break
		}
		pointList = append(pointList, p)
	}
	sort.SliceStable(pointList, func(i, j int) bool {
		return pointList[i].UUID > pointList[j].UUID
	})
	for _, p := range pointList {
		client, dev := p.Client, p.Device
		point := PointSchema{
			Uptime: p.Uptime,
			UUID: p.UUID,
			Alias: p.Alias,
			Address: client.Addr,
			Device: dev.Name(),
			RxBytes: client.Sts.RxOkay,
			TxBytes: client.Sts.TxOkay,
			ErrPkt: client.Sts.TxError,
			State: client.State(),
		}
		body.Points = append(body.Points, point)
	}

	neighborList := make([]*models.Neighbor, 0, 128)
	for n := range service.Neighbor.List() {
		if n == nil {
			break
		}
		neighborList = append(neighborList, n)
	}
	sort.SliceStable(neighborList, func(i, j int) bool {
		return neighborList[i].IpAddr.String() > neighborList[j].IpAddr.String()
	})
	for _, n := range neighborList {
		neighbor := NeighborSchema{
			Uptime: n.UpTime(),
			HwAddr: n.HwAddr.String(),
			IpAddr: n.IpAddr.String(),
			Client: n.Client.String(),
		}
		body.Neighbors = append(body.Neighbors, neighbor)
	}

	linkList := make([]*models.Point, 0, 128)
	for p := range service.Link.List() {
		if p == nil {
			break
		}
		linkList = append(linkList, p)
	}
	sort.SliceStable(linkList, func(i, j int) bool {
		return linkList[i].UUID > linkList[j].UUID
	})
	for _, p := range linkList {
		client, dev := p.Client, p.Device
		link := LinkSchema{
			UUID: p.UUID,
			Uptime:client.UpTime(),
			Device: dev.Name(),
			Address: client.Addr,
			State: client.State(),
			IpAddr: strings.Split(client.Addr, ":")[0],
		}
		body.Links = append(body.Links, link)
	}

	lineList := make([]*models.Line, 0, 128)
	for l := range service.Online.List() {
		if l == nil {
			break
		}
		lineList = append(lineList, l)
	}
	sort.SliceStable(lineList, func(i, j int) bool {
		return lineList[i].String() > lineList[j].String()
	})
	for _, l := range lineList {
		online := OnLineSchema{
			EthType: l.EthType,
			IpSource: l.IpSource.String(),
			IpDest: l.IpDest.String(),
			IpProto: libol.IpProto2Str(l.IpProtocol),
			PortSource: l.PortSource,
			PortDest: l.PortDest,
		}
		body.OnLines = append(body.OnLines, online)

	}
	return body
}

func (h *Http) IndexHtml(w http.ResponseWriter, r *http.Request) {
	body := IndexSchema{
		Points: make([]PointSchema, 0, 128),
		Links: make([]LinkSchema, 0, 128),
		Neighbors: make([]NeighborSchema, 0, 128),
		OnLines: make([]OnLineSchema, 0, 128),
	}
	file := h.getFile("/index.html")
	if t, err := template.ParseFiles(file); err == nil {
		t.Execute(w, h.getIndex(&body))
	} else {
		libol.Error("Http.Index %s", err)
		fmt.Fprintf(w, "template.ParseFiles %d", err)
	}
}

func (h *Http) ListUser(w http.ResponseWriter, r *http.Request) {
	users := make([]*models.User, 0, 1024)
	for u := range service.User.List() {
		if u == nil {
			break
		}
		users = append(users, u)
	}
	h.ResponseJson(w, users)
}

func (h *Http) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := service.User.Get(vars["id"])
	if user != nil {
		h.ResponseJson(w, user)
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h *Http) AddUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := &models.User{}
	if err := json.Unmarshal([]byte(body), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	service.User.Add(user)
	h.ResponseMsg(w, 0, "")
}

func (h *Http) DelUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("DelUser %s", vars["id"])

	service.User.Del(vars["id"])

	h.ResponseMsg(w, 0, "")
}

func (h *Http) ListNeighbor(w http.ResponseWriter, r *http.Request) {
	neighbors := make([]*models.Neighbor, 0, 1024)
	for n := range service.Neighbor.List() {
		if n == nil {
			break
		}

		neighbors = append(neighbors, n)
	}

	h.ResponseJson(w, neighbors)
}

func (h *Http) ListLink(w http.ResponseWriter, r *http.Request) {
	links := make([]*models.Point, 0, 1024)
	for l := range service.Link.List() {
		if l == nil {
			break
		}
		links = append(links, l)
	}

	h.ResponseJson(w, links)
}

func (h *Http) GetLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("GetPoint %s", vars["id"])

	link := service.Link.Get(vars["id"])
	if link != nil {
		h.ResponseJson(w, link)
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h *Http) AddLink(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := &config.Point{}
	if err := json.Unmarshal([]byte(body), c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Default()
	h.worker.AddLink(c)
	h.ResponseMsg(w, 0, "")
}

func (h *Http) DelLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	libol.Info("DelLink %s", vars["id"])

	h.worker.DelLink(vars["id"])

	h.ResponseMsg(w, 0, "")
}

func (h *Http) ListPoint(w http.ResponseWriter, r *http.Request) {
	points := make([]*models.Point, 0, 1024)
	for u := range service.Point.List() {
		if u == nil {
			break
		}
		points = append(points, u)
	}
	h.ResponseJson(w, points)
}

func (h *Http) GetPoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	point := service.Point.Get(vars["id"])
	if point != nil {
		h.ResponseJson(w, point)
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}

func (h *Http) ListNetwork(w http.ResponseWriter, r *http.Request) {
	nets := make([]*models.Network, 0, 1024)
	for u := range service.Network.List() {
		if u == nil {
			break
		}
		nets = append(nets, u)
	}
	h.ResponseJson(w, nets)
}

func (h *Http) GetNetwork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	net := service.Network.Get(vars["id"])
	if net != nil {
		h.ResponseJson(w, net)
	} else {
		http.Error(w, vars["id"], http.StatusNotFound)
	}
}
