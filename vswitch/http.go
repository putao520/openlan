package vswitch

import (
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
	wroker     *Worker
	listen     string
	adminToken string
	adminFile  string
}

func NewHttp(wroker *Worker, c *Config) (h *Http) {
	h = &Http{
		wroker:     wroker,
		listen:     c.HttpListen,
		adminToken: c.Token,
		adminFile:  c.TokenFile,
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
	libol.Debug("NewHttp on %s", h.listen)

	//hfs := http.FileServer(http.Dir("."))
	if err := http.ListenAndServe(h.listen, nil); err != nil {
		libol.Error("Http.GoStart on %s: %s", h.listen, err)
		return err
	}
	return nil
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
		body := fmt.Sprintf("# uptime: %d\n", h.wroker.UpTime())
		body += "\n"
		body += "# point accessed to this vswith.\n"
		body += "uptime, remote, device, receipt, transmis, error\n"
		for p := range h.wroker.ListPoint() {
			if p == nil {
				break
			}

			client, ifce := p.Client, p.Device
			body += fmt.Sprintf("%d, %s, %s, %d, %d, %d\n",
				client.UpTime(), client.Addr, ifce.Name(),
				client.RxOkay, client.TxOkay, client.TxError)
		}

		body += "\n"
		body += "# neighbor we discovered on this vswitch.\n"
		body += "uptime, ethernet, address, remote\n"
		for n := range h.wroker.Neighbor.ListNeighbor() {
			if n == nil {
				break
			}

			body += fmt.Sprintf("%d, %s, %s, %s\n",
				n.UpTime(), n.HwAddr, n.IpAddr, n.Client)
		}

		body += "\n"
		body += "# link which connect to other vswitch.\n"
		body += "uptime, bridge, device, remote\n"
		for p := range h.wroker.ListLink() {
			if p == nil {
				break
			}

			body += fmt.Sprintf("%d, %s, %s, %s\n",
				p.GetClient().UpTime(), p.Brname, p.Ifname, p.GetClient().Addr)
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
		for u := range h.wroker.ListUser() {
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

		h.wroker.AddUser(user)

		fmt.Fprintf(w, ApiReplyer(0, "success"))
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
		for n := range h.wroker.Neighbor.ListNeighbor() {
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
		for l := range h.wroker.ListLink() {
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
		h.wroker.AddLink(c)

		fmt.Fprintf(w, ApiReplyer(0, "success"))
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

func ApiReplyer(code int, output string) string {
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
