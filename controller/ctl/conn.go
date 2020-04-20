package ctl

import (
	"github.com/danieldin95/lightstar/libstar"
	"github.com/danieldin95/openlan-go/libol"
	"golang.org/x/net/websocket"
	"sync"
	"time"
)

type Conn struct {
	Lock   sync.RWMutex
	Conn   *websocket.Conn
	Wait   *libstar.WaitOne
	Ticker *time.Ticker
	Done   chan bool
	SendQ  chan Message
}

func (cn *Conn) Close() {
	libol.Info("Conn.Close %s", cn)
	cn.Lock.Lock()
	defer cn.Lock.Unlock()
	if cn.Conn != nil {
		cn.Conn.Close()
		cn.Conn = nil
		if cn.Wait != nil {
			cn.Wait.Done()
		}
	}
}

func (cn *Conn) Open() {
	libol.Info("Conn.Open %s", cn)
	cn.Lock.Lock()
	defer cn.Lock.Unlock()

	if cn.Ticker == nil {
		cn.Ticker = time.NewTicker(5 * time.Second)
	}
	if cn.SendQ == nil {
		cn.SendQ = make(chan Message, 1024)
	}
	if cn.Done == nil {
		cn.Done = make(chan bool)
	}
}

func (cn *Conn) once() error {
	cn.Lock.Lock()
	defer cn.Lock.Unlock()
	return nil
}

func (cn *Conn) loop() {
	libol.Info("Conn.Loop %s", cn)
	defer cn.Close()
	for {
		select {
		case m := <-cn.SendQ:
			cn.Lock.RLock()
			cn.write(m)
			cn.Lock.RUnlock()
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
	if err := Codec.Send(cn.Conn, &m); err != nil {
		libol.Error("Conn.Send %s", err)
		cn.Stop()
	}
}

func (cn *Conn) read() {
	libol.Info("Conn.Read %s", cn)
	for {
		m := &Message{}
		if cn.Conn != nil {
			err := Codec.Receive(cn.Conn, m)
			if err != nil {
				libol.Error("Conn.Read %s", err)
				break
			}
		}
		libol.Cmd("Conn.Read %s", m)
	}
	cn.Stop()
}

func (cn *Conn) Start() {
	libol.Info("Conn.Start %s", cn)
	go cn.loop()
	go cn.read()
}

func (cn *Conn) Stop() {
	libol.Info("Conn.Stop %s", cn)
	if cn.Done != nil {
		cn.Done <- true
	}
}

func (cn *Conn) Send(m Message) {
	cn.SendQ <- m
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
