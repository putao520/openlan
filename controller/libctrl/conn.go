package libctrl

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/danieldin95/openlan-go/libol"
	"golang.org/x/net/websocket"
	"strings"
	"sync"
	"time"
)

type ConnOner struct {
	Close  func(con *Conn)
	Open   func(con *Conn)
	CmdCtl func(con *Conn, m Message)
}

type Conn struct {
	Lock   sync.RWMutex
	Conn   *websocket.Conn
	Wait   *libstar.WaitOne
	Ticker *time.Ticker
	Done   chan bool
	SendQ  chan Message
	RecvQ  chan Message
	Listen *libol.SafeStrMap
	Id     string
	Oner   ConnOner
}

func (cn *Conn) Listener(name string, call Listener) {
	if cn.Listen == nil {
		cn.Listen = libol.NewSafeStrMap(32)
	}
	_ = cn.Listen.Set(strings.ToUpper(name), call)
}

func (cn *Conn) Open() {
	libol.Stack("Conn.Open %s", cn)
	cn.Lock.Lock()
	if cn.Ticker == nil {
		cn.Ticker = time.NewTicker(5 * time.Second)
	}
	if cn.SendQ == nil {
		cn.SendQ = make(chan Message, 1024)
	}
	if cn.RecvQ == nil {
		cn.RecvQ = make(chan Message, 1024)
	}
	if cn.Done == nil {
		cn.Done = make(chan bool, 2)
	}
	if cn.Listen == nil {
		cn.Listen = libol.NewSafeStrMap(32)
	}
	cn.Lock.Unlock()
	if cn.Oner.Open != nil {
		cn.Oner.Open(cn)
	}
}

func (cn *Conn) Close() {
	libol.Stack("Conn.Close %s", cn)
	if cn.Oner.Close != nil {
		cn.Oner.Close(cn)
	}
	cn.Lock.Lock()
	defer cn.Lock.Unlock()
	if cn.Conn == nil {
		return
	}
	if cn.Wait != nil {
		cn.Wait.Done()
	}
	cn.Conn.Close()
	cn.Conn = nil
	cn.SendQ = nil
	cn.RecvQ = nil
}

func (cn *Conn) dispatch(m Message) error {
	libol.Cmd("Conn.dispatch %s %s", cn.Id, &m)
	if cn.Oner.CmdCtl != nil {
		cn.Oner.CmdCtl(cn, m)
	}
	value := cn.Listen.Get(m.Resource)
	if value != nil {
		if call, ok := value.(Listener); ok {
			switch m.Action {
			case "GET":
				return call.GetCtl(cn.Id, m)
			case "ADD":
				return call.AddCtl(cn.Id, m)
			case "DEL":
				return call.DelCtl(cn.Id, m)
			case "MOD":
				return call.ModCtl(cn.Id, m)
			}
		}
		libol.Error("Conn.dispatch unknown %s", m.Resource)
	} else {
		libol.Warn("conn.dispatch notSupport %s", m.Resource)
	}
	return nil
}

func (cn *Conn) once() error {
	cn.Lock.Lock()
	defer cn.Lock.Unlock()
	return nil
}

func (cn *Conn) loop() {
	libol.Stack("Conn.Loop %s", cn)
	defer func() {
		cn.Close()
		libol.Stack("Conn.loop exit")
	}()
	for {
		select {
		case m := <-cn.SendQ:
			cn.write(m)
		case m := <-cn.RecvQ:
			if err := cn.dispatch(m); err != nil {
				libol.Error("Conn.Loop %s", err)
			}
		case <-cn.Done:
			libol.Debug("Conn.Loop %s Done", cn)
			return
		case <-cn.Ticker.C:
			if err := cn.once(); err != nil {
				libol.Error("Conn.Loop %s", err)
				return
			}
		}
	}
}

func (cn *Conn) write(m Message) {
	cn.Lock.Lock()
	defer cn.Lock.Unlock()

	libol.Cmd("Conn.write %s", m)
	if err := Codec.Send(cn.Conn, &m); err != nil {
		libol.Error("Conn.Send %s", err)
		cn.Stop()
	}
}

func (cn *Conn) read() {
	libol.Stack("Conn.read %s", cn)
	for {
		m := Message{}
		if cn.Conn != nil {
			err := Codec.Receive(cn.Conn, &m)
			if err != nil {
				libol.Error("Conn.read %s", err)
				break
			}
			libol.Cmd("Conn.Read %s", &m)
			cn.RecvQ <- m
		}
	}
	cn.Stop()
	libol.Stack("Conn.read exit")
}

func (cn *Conn) Start() {
	libol.Stack("Conn.Start %s", cn)
	go cn.loop()
	go cn.read()
}

func (cn *Conn) Stop() {
	libol.Stack("Conn.Stop %s", cn)
	if cn.Done != nil {
		cn.Done <- true
	}
}

func (cn *Conn) Send(m Message) {
	cn.Lock.RLock()
	defer cn.Lock.RUnlock()
	if cn.SendQ != nil {
		cn.SendQ <- m
	}
}

func (cn *Conn) SendWait(m Message) error {
	cn.Lock.RLock()
	defer cn.Lock.RUnlock()
	if err := Codec.Send(cn.Conn, &m); err != nil {
		return err
	}
	return nil
}

func (cn *Conn) string() string {
	if cn.Conn != nil {
		return cn.Conn.LocalAddr().String()
	}
	return ""
}

func (cn *Conn) String() string {
	cn.Lock.RLock()
	defer cn.Lock.RUnlock()
	return cn.string()
}

func (cn *Conn) Host() string {
	if cn.Conn == nil {
		return ""
	}
	if req := cn.Conn.Request(); req != nil {
		return req.Host
	}
	return ""
}

func (cn *Conn) Address() string {
	if cn.Conn == nil {
		return ""
	}
	if req := cn.Conn.Request(); req != nil {
		return req.RemoteAddr
	}
	return ""
}
