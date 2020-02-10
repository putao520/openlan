package libol

import "sync"

//m := NewSafeMap(1024)
//m.Set("hi", 1)
//a :=3
//m.Set("hip", &a)
//c := m.Get("hip").(*int)
//fmt.Printf("%s\n%d\n", m, *c)

type SafeMap struct {
	Data map[interface{}]interface{}
	Lock sync.RWMutex
}

func NewSafeMap(size int) *SafeMap {
	return &SafeMap{
		Data: make(map[interface{}]interface{}, size),
	}
}

func (sm *SafeMap) Len() int {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	return len(sm.Data)
}

func (sm *SafeMap) Set(k interface{}, v interface{}) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	sm.Data[k] = v
}

func (sm *SafeMap) Del(k interface{}) {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()

	if _, ok := sm.Data[k]; ok {
		delete(sm.Data, k)
	}
}

func (sm *SafeMap) Get(k interface{}) interface{} {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	return sm.Data[k]
}

func (sm *SafeMap) GetEx(k string) (interface{}, bool) {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	v, ok := sm.Data[k]
	return v, ok
}

func (sm *SafeMap) List() <-chan interface{} {
	ret := make(chan interface{}, 1024)

	go func() {
		sm.Lock.RLock()
		defer sm.Lock.RUnlock()

		for _, u := range sm.Data {
			ret <- u
		}
		ret <- nil //Finish channel by nil.
	}()

	return ret
}

// a := SafeVar
// a.Set(0x01)
// a.Get().(int)

type SafeVar struct {
	Data interface{}
	Lock sync.RWMutex
}

func NewSafeVar(size int) *SafeVar {
	return &SafeVar{}
}

func (sv *SafeVar) Set(v interface{}) {
	sv.Lock.Lock()
	defer sv.Lock.Unlock()
	sv.Data = v
}

func (sv *SafeVar) Get() interface{} {
	sv.Lock.RLock()
	defer sv.Lock.RUnlock()
	return sv.Data
}
