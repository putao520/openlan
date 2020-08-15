package _switch

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"io"
	"net"
	"net/http"
)

type proxy struct{}

var connectResponse = []byte("HTTP/1.1 200 OK\r\n\r\n")

func (t *proxy) Route(w http.ResponseWriter, p *http.Response) (written int64, err error) {
	for key, value := range p.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(p.StatusCode)
	return io.Copy(w, p.Body)
}

func (t *proxy) Tunnel(w http.ResponseWriter, conn net.Conn) {
	src, bio, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer src.Close()
	libol.Debug("proxy.Handle to %s", conn.RemoteAddr())
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
			libol.Debug("proxy.Handle from ws %s", err)
		}
	}()
	go func() {
		defer wait.Done()
		if _, err := io.Copy(src, conn); err != nil {
			libol.Debug("proxy.Handle from target %s", err)
		}
	}()
	wait.Wait()
	libol.Debug("proxy.Handle %s exit", conn.RemoteAddr())
}

func (t *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	libol.Debug("proxy.ServeHTTP %v", r)
	libol.Debug("proxy.ServeHTTP %v", r.URL.Host)
	if r.Method == "CONNECT" { //RFC-7231 Tunneling TCP based protocols through Web proxy servers
		conn, err := net.Dial("tcp", r.URL.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer conn.Close()
		_, _ = w.Write(connectResponse)
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
