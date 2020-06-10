package libol

import (
	"net"
	"sync"
	"time"
)

type XDP struct {
	lock     sync.RWMutex
	bufSize  int
	conn     *net.UDPConn
	addr     *net.UDPAddr
	sessions *SafeStrMap
	accept   chan *XDPConn
}

func XDPListen(addr string) (net.Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	x := &XDP{
		addr:     udpAddr,
		sessions: NewSafeStrMap(1024),
		accept:   make(chan *XDPConn, 2),
		bufSize:  MAXBUF,
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	x.conn = conn
	go x.Loop()
	return x, nil
}

// Loop forever
func (x *XDP) Loop() {
	for {
		data := make([]byte, x.bufSize)
		n, udpAddr, err := x.conn.ReadFromUDP(data)
		if err != nil {
			Error("XDP.Loop %s", err)
			break
		}
		// dispatch to XDPConn and new accept
		var newConn *XDPConn
		addr := udpAddr.String()
		obj, ok := x.sessions.GetEx(addr)
		if ok {
			newConn = obj.(*XDPConn)
		} else {
			newConn = &XDPConn{
				conn:   x.conn,
				raddr:  udpAddr,
				laddr:  x.addr,
				rqueue: make(chan []byte, 1024),
				closed: false,
			}
			_ = x.sessions.Set(addr, newConn)
			x.accept <- newConn
		}
		newConn.RxQ(data[:n])
	}
}

// Accept waits for and returns the next connection to the listener.
func (x *XDP) Accept() (net.Conn, error) {
	return <-x.accept, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (x *XDP) Close() error {
	x.lock.Lock()
	defer x.lock.Unlock()
	//if x.conn == nil {
	//	return nil
	//}
	_ = x.conn.Close()
	//x.conn = nil
	//close(x.accept)
	//x.accept = nil

	// close all conn in sessions.
	x.sessions.Iter(func(k string, v interface{}) {
		conn, ok := v.(*XDPConn)
		if ok {
			_ = conn.Close()
		}
	})
	return nil
}

// Addr returns the listener's network address.
func (x *XDP) Addr() net.Addr {
	return nil
}

type XDPConn struct {
	lock   sync.RWMutex
	conn   *net.UDPConn
	raddr  *net.UDPAddr
	laddr  *net.UDPAddr
	rqueue chan []byte
	closed bool
}

func (c *XDPConn) RxQ(b []byte) {
	c.rqueue <- b
}

func (c *XDPConn) Read(b []byte) (n int, err error) {
	d := <-c.rqueue
	return copy(b, d), nil
}

func (c *XDPConn) Write(b []byte) (n int, err error) {
	return c.conn.WriteToUDP(b, c.raddr)
}

func (c *XDPConn) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return nil
	}
	c.conn = nil
	//close(c.rqueue)
	//c.rqueue = nil
	c.closed = true

	return nil
}

func (c *XDPConn) LocalAddr() net.Addr {
	return c.laddr
}

func (c *XDPConn) RemoteAddr() net.Addr {
	return c.raddr
}

func (c *XDPConn) SetDeadline(t time.Time) error {
	Warn("XDPConn.SetDeadline %s", "implement me")
	return nil
}

func (c *XDPConn) SetReadDeadline(t time.Time) error {
	Warn("XDPConn.SetReadDeadline %s", "implement me")
	return nil
}

func (c *XDPConn) SetWriteDeadline(t time.Time) error {
	Warn("XDPConn.SetReadDeadline %s", "implement me")
	return nil
}
