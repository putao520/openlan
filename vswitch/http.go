package vswitch

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/point"
)

type Http struct {
	worker     *Worker
	listen     string
	adminToken string
	adminFile  string
	server     *http.Server
}

func NewHttp(worker *Worker, c *Config) (h *Http) {
	h = &Http{
		worker:     worker,
		listen:     c.HttpListen,
		adminToken: c.Token,
		adminFile:  c.TokenFile,
		server:	    &http.Server{Addr: c.HttpListen},
	}

	if h.adminToken == "" {
		h.LoadToken()
	}

	if h.adminToken == "" {
		h.adminToken = libol.GenToken(13)
	}

	h.SaveToken()
	http.HandleFunc("/", h.Index)
	http.HandleFunc("/hello", h.Hello)
	http.HandleFunc("/api/user", h._User)
	http.HandleFunc("/api/neighbor", h._Neighbor)
	http.HandleFunc("/api/link", h._Link)

	return
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

func (h *Http) GoStart() error {
	libol.Info("Http.GoStart %s", h.listen)

	//hfs := http.FileServer(http.Dir("."))
	if err := h.server.ListenAndServe(); err != nil {
		libol.Error("Http.GoStart on %s: %s", h.listen, err)
		return err
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

	if !ok || token != h.adminToken {
		w.Header().Set("WWW-Authenticate", "Basic")
		http.Error(w, "Authorization Required.", 401)
		return false
	}

	return true
}

func (h *Http) Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %s %q", r.Method, html.EscapeString(r.URL.Path))

	for name, headers := range r.Header {
		for _, h := range headers {
			libol.Info("Http.Hello %v: %v", name, h)
		}
	}
}

func (h *Http) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		body := fmt.Sprintf("# uptime: %d\n", h.worker.UpTime())
		body += "\n"
		body += "# point accessed to this vswith.\n"
		body += "uptime, remote, device, receipt, transmis, error, state\n"
		for p := range h.worker.ListPoint() {
			if p == nil {
				break
			}

			client, dev := p.Client, p.Device
			body += fmt.Sprintf("%d, %s, %s, %d, %d, %d, %s\n",
				client.UpTime(), client.Addr, dev.Name(),
				client.RxOkay, client.TxOkay, client.TxError, client.State())
		}

		body += "\n"
		body += "# neighbor we discovered on this vswitch.\n"
		body += "uptime, ethernet, address, remote\n"
		for n := range h.worker.Neighbor.ListNeighbor() {
			if n == nil {
				break
			}

			body += fmt.Sprintf("%d, %s, %s, %s\n",
				n.UpTime(), n.HwAddr, n.IpAddr, n.Client)
		}

		body += "\n"
		body += "# link which connect to other vswitch.\n"
		body += "uptime, bridge, device, remote, state\n"
		for p := range h.worker.ListLink() {
			if p == nil {
				break
			}
			body += fmt.Sprintf("%d, %s, %s, %s, %s\n",
				p.UpTime(), p.BrName, p.IfName(), p.Addr(), p.State())
		}

		fmt.Fprintf(w, body)
	default:
		http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
		return
	}
}

func (h *Http) Marshal(v interface{}) (string, error) {
	str, err := json.Marshal(v)
	if err != nil {
		libol.Error("Http.Marsha1: %s", err)
		return "", err
	}

	return string(str), nil
}

func (h *Http) _User(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuth(w, r) {
		return
	}

	switch r.Method {
	case "GET":
		users := make([]*User, 0, 1024)
		for u := range h.worker.ListUser() {
			if u == nil {
				break
			}
			users = append(users, u)
		}

		body, _ := h.Marshal(users)
		fmt.Fprintf(w, body)
	case "POST":
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error| Http._User: %s", err), 400)
			return
		}

		user := &User{}
		if err := json.Unmarshal([]byte(body), user); err != nil {
			http.Error(w, fmt.Sprintf("Error| Http._User: %s", err), 400)
			return
		}

		h.worker.AddUser(user)

		fmt.Fprintf(w, ApiReplier(0, "success"))
	default:
		http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
	}
}

func (h *Http) _Neighbor(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuth(w, r) {
		return
	}

	switch r.Method {
	case "GET":
		neighbors := make([]*Neighbor, 0, 1024)
		for n := range h.worker.Neighbor.ListNeighbor() {
			if n == nil {
				break
			}

			neighbors = append(neighbors, n)
		}

		body, _ := h.Marshal(neighbors)
		fmt.Fprintf(w, body)
	default:
		http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
		return
	}
}

func (h *Http) _Link(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuth(w, r) {
		return
	}

	switch r.Method {
	case "GET":
		links := make([]*point.Point, 0, 1024)
		for l := range h.worker.ListLink() {
			if l == nil {
				break
			}
			links = append(links, l)
		}
		body, _ := h.Marshal(links)
		fmt.Fprintf(w, body)
	case "POST":
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error| Http._Link: %s", err), 400)
			return
		}

		c := &point.Config{}
		if err := json.Unmarshal([]byte(body), c); err != nil {
			http.Error(w, fmt.Sprintf("Error| Http._Link: %s", err), 400)
			return
		}

		c.Default()
		h.worker.AddLink(c)

		fmt.Fprintf(w, ApiReplier(0, "success"))
	default:
		http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
	}
}

type ApiReply struct {
	Code   int
	Output string
}

func NewApiReply(code int, output string) (h *ApiReply) {
	h = &ApiReply{
		Code:   code,
		Output: output,
	}
	return
}

func ApiReplier(code int, output string) string {
	h := ApiReply{
		Code:   code,
		Output: output,
	}
	return h.String()
}

func (h *ApiReply) String() string {
	str, err := json.Marshal(h)
	if err != nil {
		libol.Error("ApiReply.String error: %s", err)
		return ""
	}
	return string(str)
}
