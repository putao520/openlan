package libol

import (
	"net"
	"sync"
	"time"
)

type XDP struct {
	lock       sync.RWMutex
	bufSize    int
	connection *net.UDPConn
	address    *net.UDPAddr
	sessions   *SafeStrMap
	accept     chan *XDPConn
}

func XDPListen(addr string) (net.Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	x := &XDP{
		address:  udpAddr,
		sessions: NewSafeStrMap(1024),
		accept:   make(chan *XDPConn, 2),
		bufSize:  MAXBUF,
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	x.connection = conn
	go x.Loop()
	return x, nil
}

// Loop forever
func (x *XDP) Loop() {
	for {
		data := make([]byte, x.bufSize)
		n, udpAddr, err := x.connection.ReadFromUDP(data)
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
				connection: x.connection,
				remoteAddr: udpAddr,
				localAddr:  x.address,
				readQueue:  make(chan []byte, 1024),
				closed:     false,
			}
			_ = x.sessions.Set(addr, newConn)
			x.accept <- newConn
		}
		newConn.toQueue(data[:n])
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

	_ = x.connection.Close()
	// close all connection in sessions.
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
	return x.address
}

type XDPConn struct {
	lock       sync.RWMutex
	connection *net.UDPConn
	remoteAddr *net.UDPAddr
	localAddr  *net.UDPAddr
	readQueue  chan []byte
	closed     bool
	readDead   time.Time
	writeDead  time.Time
}

func (c *XDPConn) toQueue(b []byte) {
	c.lock.RLock()
	if c.closed {
		c.lock.RUnlock()
		return
	} else {
		c.lock.RUnlock()
	}
	c.readQueue <- b
}

func (c *XDPConn) Read(b []byte) (n int, err error) {
	c.lock.RLock()
	if c.closed {
		c.lock.RUnlock()
		return 0, NewErr("read on closed")
	}
	var timeout *time.Timer
	outChan := make(<-chan time.Time)
	if !c.readDead.IsZero() {
		if time.Now().After(c.readDead) {
			c.lock.RUnlock()
			return 0, NewErr("read timeout")
		}
		delay := c.readDead.Sub(time.Now())
		timeout = time.NewTimer(delay)
		outChan = timeout.C
	}
	c.lock.RUnlock()

	// wait for read event or timeout or error
	select {
	case <-outChan:
		return 0, NewErr("read timeout")
	case d := <-c.readQueue:
		return copy(b, d), nil
	}
}

func (c *XDPConn) Write(b []byte) (n int, err error) {
	c.lock.RLock()
	if c.closed {
		c.lock.RUnlock()
		return 0, NewErr("write to closed")
	} else {
		c.lock.RUnlock()
	}
	return c.connection.WriteToUDP(b, c.remoteAddr)
}

func (c *XDPConn) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return nil
	}
	c.connection = nil
	c.closed = true

	return nil
}

func (c *XDPConn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *XDPConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *XDPConn) SetDeadline(t time.Time) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.readDead = t
	c.writeDead = t
	return nil
}

func (c *XDPConn) SetReadDeadline(t time.Time) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.readDead = t
	return nil
}

func (c *XDPConn) SetWriteDeadline(t time.Time) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.writeDead = t
	return nil
}
