package olsw

import (
	"encoding/base64"
	"github.com/armon/go-socks5"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type HttpProxy struct {
	users  map[string]string
	out    *libol.SubLogger
	server *http.Server
	cfg    *config.HttpProxy
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

func NewHttpProxy(cfg *config.HttpProxy) *HttpProxy {
	h := &HttpProxy{
		out: libol.NewSubLogger(cfg.Listen),
		cfg: cfg,
	}
	h.server = &http.Server{
		Addr:    cfg.Listen,
		Handler: h,
	}
	auth := cfg.Auth
	if len(auth.Username) > 0 {
		h.users = make(map[string]string, 1)
		h.users[auth.Username] = auth.Password
	}
	return h
}

func (t *HttpProxy) isAuth(username, password string) bool {
	if p, ok := t.users[username]; ok {
		return p == password
	}
	return false
}

func (t *HttpProxy) CheckAuth(w http.ResponseWriter, r *http.Request) bool {
	if len(t.users) == 0 {
		return true
	}
	auth := r.Header.Get("Proxy-Authorization")
	user, password, ok := parseBasicAuth(auth)
	if !ok || !t.isAuth(user, password) {
		w.Header().Set("Proxy-Authenticate", "Basic")
		http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
		return false
	}
	return true
}

func (t *HttpProxy) route(w http.ResponseWriter, p *http.Response) {
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
	_, _ = io.Copy(w, p.Body)
}

func (t *HttpProxy) tunnel(w http.ResponseWriter, conn net.Conn) {
	src, bio, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer src.Close()
	t.out.Info("HttpProxy.tunnel %s -> %s", src.RemoteAddr(), conn.RemoteAddr())
	wait := libol.NewWaitOne(2)
	libol.Go(func() {
		defer wait.Done()
		// The returned bufio.Reader may contain unprocessed buffered data from the client.
		// Copy them to dst so we can use src directly.
		if n := bio.Reader.Buffered(); n > 0 {
			n64, err := io.CopyN(conn, bio, int64(n))
			if n64 != int64(n) || err != nil {
				t.out.Warn("HttpProxy.tunnel io.CopyN:", n64, err)
				return
			}
		}
		if _, err := io.Copy(conn, src); err != nil {
			t.out.Debug("HttpProxy.tunnel from ws %s", err)
		}
	})
	libol.Go(func() {
		defer wait.Done()
		if _, err := io.Copy(src, conn); err != nil {
			t.out.Debug("HttpProxy.tunnel from target %s", err)
		}
	})
	wait.Wait()
	t.out.Debug("HttpProxy.tunnel %s exit", conn.RemoteAddr())
}

func (t *HttpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.out.Debug("HttpProxy.ServeHTTP %v", r)
	t.out.Debug("HttpProxy.ServeHTTP %v", r.URL.Host)
	if !t.CheckAuth(w, r) {
		t.out.Info("HttpProxy.ServeHTTP Required %v Authentication", r.URL.Host)
		return
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
		t.route(w, p)
	}
}

func (t *HttpProxy) Start() {
	if t.server == nil || t.cfg == nil {
		return
	}
	crt := t.cfg.Cert
	if crt == nil || crt.KeyFile == "" {
		t.out.Info("HttpProxy.start http://%s", t.server.Addr)
	} else {
		t.out.Info("HttpProxy.start https://%s", t.server.Addr)
	}
	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Go(func() error {
		defer t.server.Shutdown(nil)
		if crt == nil || crt.KeyFile == "" {
			if err := t.server.ListenAndServe(); err != nil {
				t.out.Warn("HttpProxy.start %s", err)
				return err
			}
		} else {
			if err := t.server.ListenAndServeTLS(crt.CrtFile, crt.KeyFile); err != nil {
				t.out.Error("HttpProxy.start %s", err)
				return err
			}
		}
		return nil
	})
}

type TcpProxy struct {
	listen   string
	target   []string
	listener net.Listener
	out      *libol.SubLogger
	rr       uint64
}

func NewTcpProxy(cfg *config.TcpProxy) *TcpProxy {
	return &TcpProxy{
		listen: cfg.Listen,
		target: cfg.Target,
		out:    libol.NewSubLogger(cfg.Listen),
	}
}

func (t *TcpProxy) tunnel(src net.Conn, dst net.Conn) {
	defer dst.Close()
	defer src.Close()
	t.out.Info("TcpProxy.tunnel %s -> %s", src.RemoteAddr(), dst.RemoteAddr())
	wait := libol.NewWaitOne(2)
	libol.Go(func() {
		defer wait.Done()
		if _, err := io.Copy(dst, src); err != nil {
			t.out.Debug("TcpProxy.tunnel from ws %s", err)
		}
	})
	libol.Go(func() {
		defer wait.Done()
		if _, err := io.Copy(src, dst); err != nil {
			t.out.Debug("TcpProxy.tunnel from target %s", err)
		}
	})
	wait.Wait()
	t.out.Debug("TcpProxy.tunnel %s exit", dst.RemoteAddr())
}

func (t *TcpProxy) loadBalance() string {
	i := t.rr % uint64(len(t.target))
	t.rr++
	return t.target[i]
}

func (t *TcpProxy) Start() {
	var listen net.Listener
	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Done(func() error {
		var err error
		listen, err = net.Listen("tcp", t.listen)
		if err != nil {
			t.out.Warn("TcpProxy.Start %s", err)
		}
		return err
	})
	t.listener = listen
	t.out.Info("TcpProxy.Start: %s", t.target)
	libol.Go(func() {
		defer listen.Close()
		for {
			conn, err := listen.Accept()
			if err != nil {
				t.out.Error("TcpServer.Accept: %s", err)
				break
			}
			// connect target and pipe it.
			backend := t.loadBalance()
			target, err := net.Dial("tcp", backend)
			if err != nil {
				t.out.Error("TcpProxy.Accept %s", err)
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
	if t.listener != nil {
		t.listener.Close()
	}
	t.out.Info("TcpProxy.Stop")
	t.listener = nil
}

type SocksProxy struct {
	server *socks5.Server
	out    *libol.SubLogger
	cfg    *config.SocksProxy
}

func NewSocksProxy(cfg *config.SocksProxy) *SocksProxy {
	s := &SocksProxy{
		cfg: cfg,
		out: libol.NewSubLogger(cfg.Listen),
	}
	// Create a SOCKS5 server
	auth := cfg.Auth
	authMethods := make([]socks5.Authenticator, 0, 2)
	if len(auth.Username) > 0 {
		author := socks5.UserPassAuthenticator{
			Credentials: socks5.StaticCredentials{
				auth.Username: auth.Password,
			},
		}
		authMethods = append(authMethods, author)
	}
	conf := &socks5.Config{AuthMethods: authMethods}
	server, err := socks5.New(conf)
	if err != nil {
		s.out.Error("NewSocksProxy %s", err)
		return nil
	}
	s.server = server
	return s
}

func (s *SocksProxy) Start() {
	if s.server == nil || s.cfg == nil {
		return
	}
	addr := s.cfg.Listen
	s.out.Info("Proxy.startSocks")

	promise := &libol.Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
	promise.Go(func() error {
		if err := s.server.ListenAndServe("tcp", addr); err != nil {
			s.out.Warn("Proxy.startSocks %s", err)
			return err
		}
		return nil
	})
}

type Proxy struct {
	cfg   *config.Proxy
	tcp   map[string]*TcpProxy
	socks map[string]*SocksProxy
	http  map[string]*HttpProxy
}

func NewProxy(cfg *config.Proxy) *Proxy {
	return &Proxy{
		cfg:   cfg,
		socks: make(map[string]*SocksProxy, 32),
		tcp:   make(map[string]*TcpProxy, 32),
		http:  make(map[string]*HttpProxy, 32),
	}
}

func (p *Proxy) Initialize() {
	if p.cfg == nil {
		return
	}
	for _, c := range p.cfg.Socks {
		s := NewSocksProxy(c)
		if s == nil {
			continue
		}
		p.socks[c.Listen] = s
	}
	for _, c := range p.cfg.Tcp {
		p.tcp[c.Listen] = NewTcpProxy(c)
	}
	for _, c := range p.cfg.Http {
		if c == nil || c.Listen == "" {
			continue
		}
		h := NewHttpProxy(c)
		p.http[c.Listen] = h
	}
}

func (p *Proxy) Start() {
	if p.cfg == nil {
		return
	}
	libol.Info("Proxy.Start")
	for _, s := range p.socks {
		s.Start()
	}
	for _, t := range p.tcp {
		t.Start()
	}
	for _, h := range p.http {
		h.Start()
	}
}

func (p *Proxy) Stop() {
	if p.cfg == nil {
		return
	}
	libol.Info("Proxy.Stop")
	for _, t := range p.tcp {
		t.Stop()
	}
}

func init() {
	// HTTP/2.0 not support upgrade for Hijacker
	if err := os.Setenv("GODEBUG", "http2server=0"); err != nil {
		libol.Warn("proxy.init %s")
	}
}
