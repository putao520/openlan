package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/libol"
	_ "github.com/lightstar-dev/openlan-go/models"
	"github.com/lightstar-dev/openlan-go/point"
	"github.com/lightstar-dev/openlan-go/vswitch"
)

type HttpResource interface {
	Get (http.ResponseWriter, *http.Request)
	Post (http.ResponseWriter, *http.Request)
	Put (http.ResponseWriter, *http.Request)
	Delete (http.ResponseWriter, *http.Request)
	Patch (http.ResponseWriter, *http.Request)
}

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
}

func (h *HttpServer) Initialize(c *config.OpenLan) {
	if h.router == nil {
		h.router = mux.NewRouter()
	}
	h.listen = c.HttpListen
	h.adminToken = c.Token
	h.adminFile = c.TokenFile
	if h.server == nil {
		h.server = &http.Server{
			Addr: c.HttpListen,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler: h.router,
		}
	}
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

func (h *HttpServer) AddResource(path string, hr HttpResource) {
	libol.Info("HttpServer.AddResource: %s on %s", path, hr)

	h.router.HandleFunc(path, hr.Get).Methods("GET")
	h.router.HandleFunc(path, hr.Post).Methods("POST")
	h.router.HandleFunc(path, hr.Put).Methods("PUT")
	h.router.HandleFunc(path, hr.Delete).Methods("Delete")
	h.router.HandleFunc(path, hr.Patch).Methods("PATCH")
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
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Authorization Required.", http.StatusUnauthorized)
		return false
	}

	return true
}

func (h *HttpServer) UpTime() int64 {
	return time.Now().Unix() - h.newTime
}

func (h *HttpServer) Marshal(v interface{}) (string, error) {
	str, err := json.Marshal(v)
	if err != nil {
		libol.Error("HttpServer.Marsha1: %s", err)
		return "", err
	}

	return string(str), nil
}

var App = &HttpServer{
	router: mux.NewRouter(),
}

type InterfaceResource struct {
}

func (ir *InterfaceResource) Get(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This Method not allowed", http.StatusMethodNotAllowed)
}

func (ir *InterfaceResource) Post(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This Method not allowed", http.StatusMethodNotAllowed)
}

func (ir *InterfaceResource) Put(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This Method not allowed", http.StatusMethodNotAllowed)
}

func (ir *InterfaceResource) Delete(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This Method not allowed", http.StatusMethodNotAllowed)
}

func (ir *InterfaceResource) Patch(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This Method not allowed", http.StatusMethodNotAllowed)
}


type HiResource struct {
	InterfaceResource
}

func (h *HiResource) Get(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi"))
}

func init() {
	App.AddResource("/hi", &HiResource{})
}

func main() {
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