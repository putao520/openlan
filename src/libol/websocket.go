package libol

import (
	"crypto/tls"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"time"
)

type wsConn struct {
	*websocket.Conn
}

func (ws *wsConn) RemoteAddr() net.Addr {
	req := ws.Request()
	if req == nil {
		return ws.RemoteAddr()
	}
	addr := req.RemoteAddr
	if ret, err := net.ResolveTCPAddr("tcp", addr); err == nil {
		return ret
	}
	return nil
}

type WebCa struct {
	CaKey string
	CaCrt string
}

type WebConfig struct {
	Ca      *WebCa
	Block   kcp.BlockCrypt
	Timeout time.Duration // ns
}

// Server Implement

type WebServer struct {
	*socketServer
	webCfg   *WebConfig
	listener *http.Server
}

func NewWebServer(listen string, cfg *WebConfig) *WebServer {
	t := &WebServer{
		webCfg:       cfg,
		socketServer: NewSocketServer(listen),
	}
	t.close = t.Close
	if err := t.Listen(); err != nil {
		Debug("NewWebServer: %s", err)
	}
	return t
}

func (t *WebServer) Listen() (err error) {
	if t.webCfg.Ca != nil {
		Info("WebServer.Listen: wss://%s", t.address)
	} else {
		Info("WebServer.Listen: ws://%s", t.address)
	}
	t.listener = &http.Server{
		Addr: t.address,
	}
	return nil
}

func (t *WebServer) Close() {
	if t.listener != nil {
		_ = t.listener.Close()
		Info("WebServer.Close: %s", t.address)
		t.listener = nil
	}
}

func (t *WebServer) Accept() {
	Debug("WebServer.Accept")
	for {
		if t.listener != nil {
			break
		}
		if err := t.Listen(); err != nil {
			Warn("WebServer.Accept: %s", err)
		}
		time.Sleep(time.Second * 5)
	}
	defer t.Close()
	t.listener.Handler = websocket.Handler(func(ws *websocket.Conn) {
		if !t.preAccept(ws) {
			return
		}
		defer ws.Close()
		ws.PayloadType = websocket.BinaryFrame
		wws := &wsConn{ws}
		client := NewWebClientFromConn(wws, t.webCfg)
		t.onClients <- client
		<-client.done
		Info("WebServer.Accept: %s exit", ws.RemoteAddr())
	})
	if t.webCfg.Ca == nil {
		if err := t.listener.ListenAndServe(); err != nil {
			Error("WebServer.Accept on %s: %s", t.address, err)
		}
	} else {
		ca := t.webCfg.Ca
		if err := t.listener.ListenAndServeTLS(ca.CaCrt, ca.CaKey); err != nil {
			Error("WebServer.Accept on %s: %s", t.address, err)
		}
	}
}

// Client Implement

type WebClient struct {
	socketClient
	webCfg *WebConfig
	done   chan bool
}

func NewWebClient(addr string, cfg *WebConfig) *WebClient {
	t := &WebClient{
		webCfg:       cfg,
		socketClient: NewSocketClient(addr, &StreamMessage{
			block: cfg.Block,
		}),
		done:         make(chan bool, 2),
	}
	t.connector = t.Connect
	return t
}

func NewWebClientFromConn(conn net.Conn, cfg *WebConfig) *WebClient {
	addr := conn.RemoteAddr().String()
	t := &WebClient{
		webCfg:       cfg,
		socketClient: NewSocketClient(addr, &StreamMessage{
			block: cfg.Block,
		}),
		done:         make(chan bool, 2),
	}
	t.updateConn(conn)
	t.connector = t.Connect
	return t
}

func (t *WebClient) Connect() error {
	if !t.retry() {
		return nil
	}
	var url string
	if t.webCfg.Ca != nil {
		Info("WebClient.Connect: wss://%s", t.address)
		url = "wss://" + t.address
	} else {
		Info("WebClient.Connect: ws://%s", t.address)
		url = "ws://" + t.address
	}
	config, err := websocket.NewConfig(url, url)
	if err != nil {
		return err
	}
	config.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	conn, err := websocket.DialConfig(config)
	if err != nil {
		return err
	}
	t.SetConnection(conn)
	if t.listener.OnConnected != nil {
		_ = t.listener.OnConnected(t)
	}
	return nil
}

func (t *WebClient) Close() {
	Info("WebClient.Close: %s %v", t.address, t.IsOk())
	t.lock.Lock()
	if t.connection != nil {
		if t.status != ClTerminal {
			t.status = ClClosed
		}
		_ = t.connection.Close()
		t.done <- true
		t.connection = nil
		t.private = nil
		t.lock.Unlock()
		if t.listener.OnClose != nil {
			_ = t.listener.OnClose(t)
		}
		Info("WebClient.Close: %s %d", t.address, t.status)
	} else {
		t.lock.Unlock()
	}
}

func (t *WebClient) Terminal() {
	t.SetStatus(ClTerminal)
	t.Close()
}

func (t *WebClient) SetStatus(v uint8) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.status != v {
		if t.listener.OnStatus != nil {
			t.listener.OnStatus(t, t.status, v)
		}
		t.status = v
	}
}
