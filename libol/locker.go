package libol

import (
	"path"
	"runtime"
	"sync"
)

type Locker struct {
	sync.Mutex
	Name string
}

func NewLocker(name string) Locker {
	return Locker{
		Mutex: sync.Mutex{},
		Name:  name,
	}
}

func (l *Locker) Lock() {
	if pc, _, line, ok := runtime.Caller(1); ok {
		name := runtime.FuncForPC(pc).Name()
		Lock("Locker.Lock: %s:%d", path.Base(name), line)
	} else {
		Lock("Locker.Lock: %s", l.Name)
	}
	l.Mutex.Lock()
}

func (l *Locker) Unlock() {
	if pc, _, line, ok := runtime.Caller(1); ok {
		name := runtime.FuncForPC(pc).Name()
		Lock("Locker.Unlock: %s:%d", path.Base(name), line)
	} else {
		Lock("Locker.Unlock: %s", l.Name)
	}
	l.Mutex.Unlock()
}
