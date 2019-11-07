package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/libol"
	_ "github.com/lightstar-dev/openlan-go/models"
	"github.com/lightstar-dev/openlan-go/point"
	"github.com/lightstar-dev/openlan-go/vswitch"
)

type HttpServer struct {
	listen     string
	adminToken string
	adminFile  string
	server     *http.Server
	crtFile    string
	keyFile    string
	pubDir     string
	newTime    int64
	switchs    map[string]*vswitch.VSwitch
	points     map[string]*point.Point
	router     *mux.Router
	lock       sync.RWMutex
}

func (h *HttpServer) Initialize(c *config.OpenLan) {
	r := h.GetRouter()
	if h.server == nil {
		h.server = &http.Server{
			Addr:         c.HttpListen,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      r,
		}
	}

	h.listen = c.HttpListen
	h.adminToken = c.Token
	h.adminFile = c.TokenFile
	h.crtFile = c.CrtFile
	h.keyFile = c.KeyFile
	h.pubDir = c.HttpDir
	h.newTime = time.Now().Unix()

	h.switchs = make(map[string]*vswitch.VSwitch, 1024)
	h.points = make(map[string]*point.Point, 1024)

	if h.adminToken == "" {
		h.LoadToken()
	}
	if h.adminToken == "" {
		h.adminToken = libol.GenToken(13)
	}

	h.SaveToken()
}

func (h *HttpServer) LoadRouter() {
	h.GetRouter().HandleFunc("/", h.GetHi).Methods("GET")
	h.GetRouter().HandleFunc("/app", h.GetApp).Methods("GET")
	h.GetRouter().HandleFunc("/point", h.GetPoint).Methods("GET")
	h.GetRouter().HandleFunc("/point/{id}", h.GetPoint).Methods("GET")
	h.GetRouter().HandleFunc("/point/{id}", h.PostPoint).Methods("POST").
		HeadersRegexp("Content-Type", "application/json")
	h.GetRouter().HandleFunc("/point/{id}", h.DeletePoint).Methods("DELETE")
	h.GetRouter().HandleFunc("/switch", h.GetSwitch).Methods("GET")
	h.GetRouter().HandleFunc("/switch/{id}", h.GetSwitch).Methods("GET")
	h.GetRouter().HandleFunc("/switch/{id}", h.PostSwitch).Methods("POST").
		HeadersRegexp("Content-Type", "application/json")
	h.GetRouter().HandleFunc("/switch/{id}", h.DeleteSwitch).Methods("DELETE")
	h.GetRouter().HandleFunc("/switch/{id}/link", h.GetSwitch).Methods("GET")
	h.GetRouter().HandleFunc("/switch/{id}/link", h.GetHi).Methods("POST").
		HeadersRegexp("Content-Type", "application/json")
	h.GetRouter().HandleFunc("/switch/{id}/link", h.GetHi).Methods("DELETE")
	h.GetRouter().HandleFunc("/switch/{id}/neighbor", h.GetSwitch).Methods("GET")
	h.GetRouter().HandleFunc("/switch/{id}/point", h.GetHi).Methods("GET")
	h.GetRouter().HandleFunc("/switch/{id}/online", h.GetHi).Methods("GET")
}

func (h *HttpServer) SaveToken() error {
	libol.Info("HttpServer.SaveToken: AdminToken: %s", h.adminToken)

	f, err := os.OpenFile(h.adminFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	defer f.Close()
	if err != nil {
		libol.Error("HttpServer.SaveToken: %s", err)
		return err
	}

	if _, err := f.Write([]byte(h.adminToken)); err != nil {
		libol.Error("HttpServer.SaveToken: %s", err)
		return err
	}

	return nil
}

func (h *HttpServer) LoadToken() error {
	if _, err := os.Stat(h.adminFile); os.IsNotExist(err) {
		libol.Info("HttpServer.LoadToken: file:%s does not exist", h.adminFile)
		return nil
	}

	contents, err := ioutil.ReadFile(h.adminFile)
	if err != nil {
		libol.Error("HttpServer.LoadToken: file:%s %s", h.adminFile, err)
		return err

	}

	h.adminToken = string(contents)
	return nil
}

func (h *HttpServer) GoStart() error {
	libol.Info("HttpServer.GoStart %s", h.listen)

	h.LoadRouter()
	if h.keyFile == "" || h.crtFile == "" {
		if err := h.server.ListenAndServe(); err != nil {
			libol.Error("HttpServer.GoStart on %s: %s", h.listen, err)
			return err
		}
	} else {
		if err := h.server.ListenAndServeTLS(h.crtFile, h.keyFile); err != nil {
			libol.Error("HttpServer.GoStart on %s: %s", h.listen, err)
			return err
		}
	}
	return nil
}

func (h *HttpServer) Shutdown() {
	libol.Info("HttpServer.Shutdown %s", h.listen)
	if err := h.server.Shutdown(context.Background()); err != nil {
		// Error from closing listeners, or context timeout:
		libol.Error("HttpServer.Shutdown: %v", err)
	}
}

func (h *HttpServer) IsAuth(w http.ResponseWriter, r *http.Request) bool {
	token, pass, ok := r.BasicAuth()
	libol.Debug("HttpServer.IsAuth token: %s, pass: %s", token, pass)

	if !ok || token != h.adminToken {
		return false
	}

	return true
}

func (h *HttpServer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.IsAuth(w, r) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic")
			http.Error(w, "Authorization Required.", http.StatusUnauthorized)
		}
	})
}

func (h *HttpServer) UpTime() int64 {
	return time.Now().Unix() - h.newTime
}

func (h *HttpServer) GetRouter() *mux.Router {
	if h.router == nil {
		h.router = mux.NewRouter()
		h.router.Use(h.Middleware)
	}

	return h.router
}

func (h *HttpServer) listPoint() <-chan *point.Point {
	c := make(chan *point.Point, 128)

	go func() {
		h.lock.RLock()
		defer h.lock.RUnlock()

		for _, p := range h.points {
			c <- p
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (h *HttpServer) getPoint(addr string) *point.Point {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return h.points[addr]
}

func (h *HttpServer) addPoint(addr string, c *config.Point) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.points[addr]; !ok {
		p := point.NewPoint(c)
		h.points[addr] = p

		go p.Start()
	}

	return nil
}

func (h *HttpServer) delPoint(addr string) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if p, ok := h.points[addr]; ok {
		p.Stop()
		delete(h.points, addr)
	}

	return nil
}

func (h *HttpServer) listSwitch() <-chan *vswitch.VSwitch {
	c := make(chan *vswitch.VSwitch, 128)

	go func() {
		h.lock.RLock()
		defer h.lock.RUnlock()

		for _, s := range h.switchs {
			c <- s
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (h *HttpServer) getSwitch(addr string) *vswitch.VSwitch {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return h.switchs[addr]
}

func (h *HttpServer) addSwitch(addr string, c *config.VSwitch) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.switchs[addr]; !ok {
		s := vswitch.NewVSwitch(c)
		h.switchs[addr] = s

		s.Start()
	}

	return nil
}

func (h *HttpServer) delSwitch(addr string) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if s, ok := h.switchs[addr]; ok {
		s.Stop()
		delete(h.switchs, addr)
	}

	return nil
}

func (h *HttpServer) Response(w http.ResponseWriter, code int, message string) {
	ret := struct {
		Code    int
		Message string
	}{
		Code:    code,
		Message: message,
	}
	h.ResponseJson(w, ret)
}

func (h *HttpServer) ResponseJson(w http.ResponseWriter, v interface{}) {
	str, err := json.Marshal(v)
	if err == nil {
		w.Write(str)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *HttpServer) GetHi(w http.ResponseWriter, r *http.Request) {
	h.Response(w, 0, fmt.Sprintf("Welcome to OpenLan by <%s>!", r.URL.Path))
}

func (h *HttpServer) GetApp(w http.ResponseWriter, r *http.Request) {
	type App struct {
		UpTime int64
	}

	app := App{UpTime: h.UpTime()}
	h.ResponseJson(w, app)
}

func (h *HttpServer) GetPoint(w http.ResponseWriter, r *http.Request) {
	type Point struct {
		UpTime int64
		BrName string `json:",omitempty"`
		IfName string
		Remote string
		State  string
	}
	vars := mux.Vars(r)
	libol.Info("GetPoint %s", vars["id"])

	if id, ok := vars["id"]; ok {
		p := h.getPoint(id)
		if p == nil {
			http.Error(w, id, http.StatusNotFound)
			return
		} else {
			data := Point{
				UpTime: p.UpTime(),
				BrName: p.BrName,
				IfName: p.IfName(),
				Remote: p.Addr(),
				State:  p.State(),
			}
			h.ResponseJson(w, data)
		}
		return
	}

	points := make([]interface{}, 0, 32)
	for p := range h.listPoint() {
		if p == nil {
			break
		}

		data := Point{
			UpTime: p.UpTime(),
			BrName: p.BrName,
			IfName: p.IfName(),
			Remote: p.Addr(),
			State:  p.State(),
		}
		points = append(points, data)
	}
	h.ResponseJson(w, points)
}

func (h *HttpServer) PostPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	id := vars["id"]
	libol.Info("PostPoint %s", id)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := config.PointDefault
	if err := json.Unmarshal([]byte(body), &c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Addr = id
	c.Right()

	h.addPoint(id, &c)
	h.Response(w, 0, "success")
}

func (h *HttpServer) DeletePoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	libol.Info("DeletePoint %s", id)

	h.delPoint(id)
	h.Response(w, 0, "success")
}

func (h *HttpServer) GetSwitch(w http.ResponseWriter, r *http.Request) {
	//type Point struct {
	//	UpTime int64
	//	BrName string `json:",omitempty"`
	//	IfName string
	//	Remote string
	//	State  string
	//}
	type Switch struct {
		UpTime int64
		//Links  []*Point
		//Points []*Point
		State   string
		Address string
	}

	vars := mux.Vars(r)
	libol.Info("GetSwitch %s", vars["id"])

	if id, ok := vars["id"]; ok {
		s := h.getSwitch(id)
		if s == nil {
			http.Error(w, id, http.StatusNotFound)
			return
		} else {
			data := Switch{
				UpTime:  s.GetWorker().UpTime(),
				State:   s.GetState(),
				Address: id,
			}
			h.ResponseJson(w, data)
		}
		return
	}

	switchs := make([]interface{}, 0, 32)
	for s := range h.listSwitch() {
		if s == nil {
			break
		}

		data := Switch{
			UpTime:  s.GetWorker().UpTime(),
			State:   s.GetState(),
			Address: s.GetWorker().GetId(),
		}
		switchs = append(switchs, data)
	}
	h.ResponseJson(w, switchs)
}

func (h *HttpServer) PostSwitch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	id := vars["id"]
	libol.Info("PostPoint %s", id)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := config.VSwitchDefault
	if err := json.Unmarshal([]byte(body), &c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.TcpListen = id
	c.Right()

	h.addSwitch(id, &c)
	h.Response(w, 0, "success")
}

func (h *HttpServer) DeleteSwitch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	libol.Info("DeleteSwitch %s", id)

	h.delSwitch(id)
	h.Response(w, 0, "success")
}

func main() {
	var App = &HttpServer{}

	c := config.NewOpenLan()
	App.Initialize(c)

	go App.GoStart()

	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-x
		App.Shutdown()
		fmt.Println("Done!")
		os.Exit(0)
	}()

	for {
		time.Sleep(1000 * time.Second)
	}
}
