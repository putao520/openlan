package libol

import (
	"os"
	"path"
	"reflect"
	"runtime"
	"runtime/pprof"
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
	ptr := reflect.ValueOf(call).Pointer()
	name := runtime.FuncForPC(ptr).Name()
	go func() {
		defer Catch("Go.func")
		Gos.Add(call)
		Info("Go.Add: %s", path.Base(name))
		call()
		Info("Go.Del: %s", path.Base(name))
		Gos.Del(call)
	}()
}

type Prof struct {
	File string
}

func (p *Prof) Start() {
	f, err := os.Create(p.File)
	if err != nil {
		Warn("Prof.Start %s", err)
		return
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		Warn("Prof.Start %s", err)
	}
}

func (p *Prof) Stop() {
	pprof.StopCPUProfile()
}
