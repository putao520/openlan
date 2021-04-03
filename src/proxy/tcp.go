package proxy

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"io"
	"net"
	"time"
)

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
