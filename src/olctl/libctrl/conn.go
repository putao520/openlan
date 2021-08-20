package libctrl

import (
	"github.com/danieldin95/openlan/src/libol"
	"golang.org/x/net/websocket"
	"strings"
	"sync"
	"time"
)

// Callback process from connection
type ConnCaller struct {
	Close  func(con *CtrlConn)
	Open   func(con *CtrlConn)
	CmdCtl func(con *CtrlConn, m Message)
	Ticker func(con *CtrlConn)
}

type ConnStats struct {
	Get    int64 `json:"get"`
	Add    int64 `json:"add"`
	Del    int64 `json:"del"`
	Mod    int64 `json:"mod"`
	Write  int64 `json:"write"`  // Write to socket buffer.
	Send   int64 `json:"send"`   // Send into Queue.
	Recv   int64 `json:"recv"`   // Receive from socket buffer.
	Action int64 `json:"action"` // Already action dispatch.
	Drop   int64 `json:"drop"`   // Dropped when Queue is fully.
}

type CtrlConn struct {
	Lock    sync.RWMutex      `json:"-"`
	Conn    *websocket.Conn   `json:"-"`
	Wait    *libol.WaitOne    `json:"-"`
	Ticker  *time.Ticker      `json:"-"`
	Done    chan bool         `json:"-"`
	SendC   int               `json:"sendC"` // Count for Queue and avoid blocking.
	SendQ   chan Message      `json:"-"`
	RecvQ   chan Message      `json:"-"`
	Listen  *libol.SafeStrMap `json:"-"`
	Id      string            `json:"id"`
	Caller  ConnCaller        `json:"-"`
	Timeout time.Duration     `json:"timeout"`
	Sts     ConnStats         `json:"statistics"`
}

func (cn *CtrlConn) Listener(name string, call Listener) {
	if cn.Listen == nil {
		cn.Listen = libol.NewSafeStrMap(32)
	}
	_ = cn.Listen.Set(strings.ToUpper(name), call)
}

func (cn *CtrlConn) Open() {
	libol.Stack("CtrlConn.Open %s", cn.Id)
	cn.Lock.Lock()
	if cn.Ticker == nil {
		cn.Ticker = time.NewTicker(5 * time.Second)
	}
	cn.SendC = 0
	cn.SendQ = make(chan Message, 1024)
	cn.RecvQ = make(chan Message, 1024)
	cn.Done = make(chan bool, 2)
	if cn.Listen == nil {
		cn.Listen = libol.NewSafeStrMap(32)
	}
	cn.Lock.Unlock()
	if cn.Caller.Open != nil {
		cn.Caller.Open(cn)
	}
}

func (cn *CtrlConn) Close() {
	libol.Info("CtrlConn.Close: %s", cn.Id)
	if cn.Caller.Close != nil {
		cn.Caller.Close(cn)
	}
	cn.Lock.Lock()
	defer cn.Lock.Unlock()
	if cn.Conn == nil {
		return
	}
	if cn.Wait != nil {
		cn.Wait.Done()
	}
	_ = cn.Conn.Close()
	cn.Conn = nil
	// close sendQ
	close(cn.SendQ)
	cn.SendQ = nil
	libol.Info("CtrlConn.Close: conn is null")
}

func (cn *CtrlConn) dispatch(m Message) error {
	libol.Cmd("CtrlConn.dispatch %s %s", cn.Id, &m)
	if cn.Caller.CmdCtl != nil {
		cn.Caller.CmdCtl(cn, m)
	}
	value := cn.Listen.Get(m.Resource)
	if value == nil {
		libol.Debug("CtrlConn.dispatch: noCall %s", m.Resource)
		return nil
	}
	if call, ok := value.(Listener); ok {
		switch m.Action {
		case "GET":
			cn.Sts.Get++
			return call.GetCtl(cn.Id, m)
		case "ADD":
			cn.Sts.Add++
			return call.AddCtl(cn.Id, m)
		case "DEL":
			cn.Sts.Del++
			return call.DelCtl(cn.Id, m)
		case "MOD":
			cn.Sts.Mod++
			return call.ModCtl(cn.Id, m)
		default:
			libol.Error("CtrlConn.dispatch: noOpr %s", m.Resource)
		}
	}
	return nil
}

func (cn *CtrlConn) once() error {
	cn.Lock.Lock()
	defer cn.Lock.Unlock()
	return nil
}

func (cn *CtrlConn) queue() {
	libol.Stack("CtrlConn.Loop %s", cn.Id)
	defer func() {
		libol.Warn("CtrlConn.queue exit")
	}()
	for {
		select {
		case m := <-cn.SendQ:
			if err := cn.write(m); err != nil {
				libol.Error("CtrlConn.queue: write %s", err)
				return
			}
			// to keep SendC is consistent, require lock
			cn.Lock.Lock()
			cn.SendC--
			cn.Sts.Write++
			cn.Lock.Unlock()
		case m := <-cn.RecvQ:
			// to avoid require lock from caller, no read lock.
			if err := cn.dispatch(m); err != nil {
				libol.Error("CtrlConn.queue: %s", err)
			}
			cn.Sts.Action++
		}
	}
}

func (cn *CtrlConn) loop() {
	libol.Stack("CtrlConn.Loop %s", cn.Id)
	defer func() {
		cn.Close()
		libol.Warn("CtrlConn.loop exit")
	}()
	for {
		select {
		case <-cn.Done:
			libol.Warn("CtrlConn.loop: recv done")
			return
		case <-cn.Ticker.C:
			if err := cn.once(); err != nil {
				libol.Error("CtrlConn.loop %s", err)
				return
			}
			if cn.Caller.Ticker != nil {
				cn.Caller.Ticker(cn)
			}
		}
	}
}

func (cn *CtrlConn) write(m Message) error {
	if cn.Conn == nil {
		return libol.NewErr("conn is null")
	}
	libol.Cmd("CtrlConn.write %s", m)
	if cn.Timeout != 0 {
		dt := time.Now().Add(cn.Timeout)
		if err := cn.Conn.SetWriteDeadline(dt); err != nil {
			return err
		}
	}
	if err := Codec.Send(cn.Conn, &m); err != nil {
		return err
	}
	return nil
}

func (cn *CtrlConn) read() {
	libol.Stack("CtrlConn.read %s", cn.Id)
	for {
		m := Message{}
		// read message from socket, no require lock.
		if cn.Conn != nil {
			if cn.Timeout != 0 {
				dt := time.Now().Add(cn.Timeout)
				if err := cn.Conn.SetReadDeadline(dt); err != nil {
					break
				}
			}
			err := Codec.Receive(cn.Conn, &m)
			if err != nil {
				libol.Error("CtrlConn.read %s", err)
				break
			}
			cn.Sts.Recv++
			libol.Cmd("CtrlConn.Read %s", &m)
			if cn.RecvQ != nil {
				cn.RecvQ <- m
			}
		}
	}
	cn.Stop()
	libol.Warn("CtrlConn.read exit")
}

func (cn *CtrlConn) Start() {
	libol.Info("CtrlConn.Start %s", cn.Id)
	libol.Go(cn.loop)
	libol.Go(cn.queue)
	libol.Go(cn.read)
}

func (cn *CtrlConn) Stop() {
	libol.Info("CtrlConn.Stop %s", cn.Id)
	if cn.Done != nil {
		cn.Done <- true
	}
}

func (cn *CtrlConn) Send(m Message) {
	cn.Lock.Lock()
	defer cn.Lock.Unlock()
	if cn.SendQ == nil {
		return
	}
	if cn.SendC >= 1024 {
		cn.Sts.Drop++
		libol.Debug("CtrlConn.Send: queue already fully")
		return
	}
	cn.SendC++
	cn.Sts.Send++
	cn.SendQ <- m
}

func (cn *CtrlConn) SendWait(m Message) error {
	cn.Lock.Lock()
	defer cn.Lock.Unlock()
	err := cn.write(m)
	cn.Sts.Write++
	return err
}

func (cn *CtrlConn) String() string {
	cn.Lock.RLock()
	defer cn.Lock.RUnlock()
	if cn.Conn != nil {
		return cn.Conn.LocalAddr().String()
	}
	return ""
}

func (cn *CtrlConn) Host() string {
	cn.Lock.RLock()
	defer cn.Lock.RUnlock()
	if cn.Conn == nil {
		return ""
	}
	if req := cn.Conn.Request(); req != nil {
		return req.Host
	}
	return ""
}

func (cn *CtrlConn) Address() string {
	cn.Lock.RLock()
	defer cn.Lock.RUnlock()
	if cn.Conn == nil {
		return ""
	}
	if req := cn.Conn.Request(); req != nil {
		return req.RemoteAddr
	}
	return ""
}
