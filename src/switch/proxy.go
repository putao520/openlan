package _switch

import (
	"encoding/base64"
	"github.com/danieldin95/openlan-go/src/libol"
	"io"
	"net"
	"net/http"
	"strings"
)

type Proxy struct {
	Users map[string]string
}

var (
	connectOkay = []byte("HTTP/1.1 200 Connection established\r\n\r\n")
)

func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
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

func (t *Proxy) isAuth(username, password string) bool {
	if p, ok := t.Users[username]; ok {
		return p == password
	}
	return false
}

func (t *Proxy) Route(w http.ResponseWriter, p *http.Response) (written int64, err error) {
	for key, value := range p.Header {
		if key == "Proxy-Authorization" {
			if len(value) > 0 { // Pop first value for next proxy.
				value = value[1:]
			}
		}
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(p.StatusCode)
	return io.Copy(w, p.Body)
}

func (t *Proxy) Tunnel(w http.ResponseWriter, conn net.Conn) {
	src, bio, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer src.Close()
	libol.Debug("Proxy.Handle to %s", conn.RemoteAddr())
	wait := libol.NewWaitOne(2)
	go func() {
		defer wait.Done()
		// The returned bufio.Reader may contain unprocessed buffered data from the client.
		// Copy them to dst so we can use src directly.
		if n := bio.Reader.Buffered(); n > 0 {
			n64, err := io.CopyN(conn, bio, int64(n))
			if n64 != int64(n) || err != nil {
				libol.Warn("io.CopyN:", n64, err)
				return
			}
		}
		if _, err := io.Copy(conn, src); err != nil {
			libol.Debug("Proxy.Handle from ws %s", err)
		}
	}()
	go func() {
		defer wait.Done()
		if _, err := io.Copy(src, conn); err != nil {
			libol.Debug("Proxy.Handle from target %s", err)
		}
	}()
	wait.Wait()
	libol.Debug("Proxy.Handle %s exit", conn.RemoteAddr())
}

func (t *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	libol.Debug("Proxy.ServeHTTP %v", r)
	libol.Debug("Proxy.ServeHTTP %v", r.URL.Host)
	if len(t.Users) != 0 {
		auth := r.Header.Get("Proxy-Authorization")
		user, password, ok := parseBasicAuth(auth)
		if !ok || !t.isAuth(user, password) {
			w.Header().Set("Proxy-Authenticate", "Basic")
			http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
			return
		}
	}
	if r.Method == "CONNECT" { //RFC-7231 Tunneling TCP based protocols through Web Proxy servers
		conn, err := net.Dial("tcp", r.URL.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer conn.Close()
		_, _ = w.Write(connectOkay)
		t.Tunnel(w, conn)
	} else { //RFC 7230 - HTTP/1.1: Message Syntax and Routing
		transport := &http.Transport{}
		p, err := transport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer transport.CloseIdleConnections()
		defer p.Body.Close()
		_, _ = t.Route(w, p)
	}
}
