package _switch

import (
	"encoding/base64"
	"github.com/armon/go-socks5"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"io"
	"net"
	"net/http"
	"strings"
)

type HttpProxy struct {
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

func (t *HttpProxy) isAuth(username, password string) bool {
	if p, ok := t.Users[username]; ok {
		return p == password
	}
	return false
}

func (t *HttpProxy) route(w http.ResponseWriter, p *http.Response) (written int64, err error) {
	defer p.Body.Close()
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

func (t *HttpProxy) tunnel(w http.ResponseWriter, conn net.Conn) {
	src, bio, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer src.Close()
	libol.Info("HttpProxy.tunnel to %s", conn.RemoteAddr())
	wait := libol.NewWaitOne(2)
	libol.Go(func() {
		defer wait.Done()
		// The returned bufio.Reader may contain unprocessed buffered data from the client.
		// Copy them to dst so we can use src directly.
		if n := bio.Reader.Buffered(); n > 0 {
			n64, err := io.CopyN(conn, bio, int64(n))
			if n64 != int64(n) || err != nil {
				libol.Warn("HttpProxy.tunnel io.CopyN:", n64, err)
				return
			}
		}
		if _, err := io.Copy(conn, src); err != nil {
			libol.Debug("HttpProxy.tunnel from ws %s", err)
		}
	})
	libol.Go(func() {
		defer wait.Done()
		if _, err := io.Copy(src, conn); err != nil {
			libol.Debug("HttpProxy.tunnel from target %s", err)
		}
	})
	wait.Wait()
	libol.Debug("HttpProxy.tunnel %s exit", conn.RemoteAddr())
}

func (t *HttpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	libol.Debug("HttpProxy.ServeHTTP %v", r)
	libol.Debug("HttpProxy.ServeHTTP %v", r.URL.Host)
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
		_, _ = w.Write(connectOkay)
		t.tunnel(w, conn)
	} else { //RFC 7230 - HTTP/1.1: Message Syntax and Routing
		transport := &http.Transport{}
		p, err := transport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer transport.CloseIdleConnections()
		_, _ = t.route(w, p)
	}
}

type TcpProxy struct {
	Listen   string
	Target   string
	Listener net.Listener
}

func (t *TcpProxy) tunnel(src net.Conn, dst net.Conn) {
	defer dst.Close()
	defer src.Close()
	libol.Info("TcpProxy.tunnel [%s,%s]->[%s,%s]",
		src.RemoteAddr(), src.LocalAddr(), dst.LocalAddr(), dst.RemoteAddr())
	wait := libol.NewWaitOne(2)
	libol.Go(func() {
		defer wait.Done()
		if _, err := io.Copy(dst, src); err != nil {
			libol.Debug("TcpProxy.tunnel from ws %s", err)
		}
	})
	libol.Go(func() {
		defer wait.Done()
		if _, err := io.Copy(src, dst); err != nil {
			libol.Debug("TcpProxy.tunnel from target %s", err)
		}
	})
	wait.Wait()
	libol.Debug("TcpProxy.tunnel %s exit", dst.RemoteAddr())
}

func (t *TcpProxy) Start() {
	listen, err := net.Listen("tcp", t.Listen)
	if err != nil {
		libol.Error("TcpProxy.Start %s", err)
		return
	}
	t.Listener = listen
	libol.Info("TcpProxy.Start: %s->%s", t.Listen, t.Target)
	libol.Go(func() {
		defer listen.Close()
		for {
			conn, err := listen.Accept()
			if err != nil {
				libol.Error("TcpServer.Accept: %s", err)
				break
			}
			// connect target and pipe it.
			target, err := net.Dial("tcp", t.Target)
			if err != nil {
				libol.Error("TcpProxy.Accept %s", err)
				continue
			}
			libol.Go(func() {
				t.tunnel(conn, target)
			})
		}
	})
	return
}

func (t *TcpProxy) Stop() {
	if t.Listener != nil {
		t.Listener.Close()
	}
	libol.Info("TcpProxy.Stop %s", t.Listen)
	t.Listener = nil
}

type Proxy struct {
	cfg   *config.Proxy
	tcp   []*TcpProxy
	socks *socks5.Server
	http  *http.Server
}

func NewProxy(cfg *config.Proxy) *Proxy {
	return &Proxy{
		cfg: cfg,
	}
}

func (p *Proxy) initSocks() {
	if p.cfg.Socks == nil || p.cfg.Socks.Listen == "" {
		return
	}
	// Create a SOCKS5 server
	auth := p.cfg.Socks.Auth
	authMethods := make([]socks5.Authenticator, 0, 2)
	if len(auth.Username) > 0 {
		author := socks5.UserPassAuthenticator{
			Credentials: socks5.StaticCredentials{
				auth.Username: auth.Password,
			},
		}
		authMethods = append(authMethods, author)
	}
	conf := &socks5.Config{
		AuthMethods: authMethods,
	}
	server, err := socks5.New(conf)
	if err != nil {
		libol.Error("Proxy.initSocks %s", err)
		return
	}
	p.socks = server
}

func (p *Proxy) initHttp() {
	if p.cfg.Http == nil || p.cfg.Http.Listen == "" {
		return
	}
	addr := p.cfg.Http.Listen
	auth := p.cfg.Http.Auth
	hp := &HttpProxy{}
	if len(auth.Username) > 0 {
		hp.Users = make(map[string]string, 1)
		hp.Users[auth.Username] = auth.Password
	}
	p.http = &http.Server{
		Addr:    addr,
		Handler: hp,
	}
}

func (p *Proxy) initTcp() {
	if len(p.cfg.Tcp) == 0 {
		return
	}
	p.tcp = make([]*TcpProxy, 0, 32)
	for _, c := range p.cfg.Tcp {
		p.tcp = append(p.tcp, &TcpProxy{
			Listen: c.Listen,
			Target: c.Target,
		})
	}
}

func (p *Proxy) startSocks() {
	if p.cfg.Socks == nil || p.socks == nil {
		return
	}
	addr := p.cfg.Socks.Listen
	libol.Info("Proxy.startSocks %s", addr)
	libol.Go(func() {
		if err := p.socks.ListenAndServe("tcp", addr); err != nil {
			libol.Error("Proxy.startSocks %s", err)
			return
		}
	})
}

func (p *Proxy) startTcp() {
	if len(p.tcp) == 0 {
		return
	}
	for _, t := range p.tcp {
		t.Start()
	}
}

func (p *Proxy) Initialize() {
	if p.cfg == nil {
		return
	}
	p.initSocks()
	p.initHttp()
	p.initTcp()
}

func (p *Proxy) startHttp() {
	if p.http == nil {
		return
	}
	libol.Info("Proxy.startHttp %s", p.http.Addr)
	libol.Go(func() {
		defer p.http.Shutdown(nil)
		if err := p.http.ListenAndServe(); err != nil {
			libol.Error("Proxy.startHttp %s", err)
		}
	})
}

func (p *Proxy) Start() {
	if p.cfg == nil {
		return
	}
	libol.Info("Proxy.Start")
	p.startTcp()
	p.startSocks()
	p.startHttp()
}

func (p *Proxy) stopTcp() {
	if len(p.tcp) == 0 {
		return
	}
	for _, t := range p.tcp {
		t.Stop()
	}
}

func (p *Proxy) Stop() {
	if p.cfg == nil {
		return
	}
	libol.Info("Proxy.Stop")
	p.stopTcp()
}
