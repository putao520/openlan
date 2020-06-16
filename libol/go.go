package libol

import (
	"path"
	"runtime"
)

type gos struct {
	lock  Locker
	total uint64
}

var Gos = gos{}

func (t *gos) Add(call interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.total++
	Debug("gos.Add %d %p", t.total, call)
}

func (t *gos) Del(call interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.total--
	Debug("gos.Del %d %p", t.total, call)
}

func Go(call func()) {
	name := "Go"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		name = runtime.FuncForPC(pc).Name()
	}
	go func() {
		Gos.Add(call)
		Info("Go.Add: %s:%d", path.Base(name), line)
		call()
		Info("Go.Del: %s:%d", path.Base(name), line)
		Gos.Del(call)
	}()
}
