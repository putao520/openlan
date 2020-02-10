package libol

import "sync"

//m := NewSafeMap(1024)
//m.Set("hi", 1)
//a :=3
//m.Set("hip", &a)
//c := m.Get("hip").(*int)
//fmt.Printf("%s\n%d\n", m, *c)

type SafeMap struct {
	size int
	data map[interface{}]interface{}
	lock sync.RWMutex
}

func NewSafeMap(size int) *SafeMap {
	calSize := size
	if calSize == 0 {
		calSize = 128
	}
	return &SafeMap{
		size: size,
		data: make(map[interface{}]interface{}, calSize),
	}
}

func (sm *SafeMap) Len() int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return len(sm.data)
}

func (sm *SafeMap) Set(k interface{}, v interface{}) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if sm.size != 0 && len(sm.data) >= sm.size {
		return NewErr("SageMap.Set already full")
	}
	sm.data[k] = v
	return nil
}

func (sm *SafeMap) Del(k interface{}) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if _, ok := sm.data[k]; ok {
		delete(sm.data, k)
	}
}

func (sm *SafeMap) Get(k interface{}) interface{} {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return sm.data[k]
}

func (sm *SafeMap) GetEx(k interface{}) (interface{}, bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	v, ok := sm.data[k]
	return v, ok
}

func (sm *SafeMap) Iter(proc func(k interface{}, v interface{})) int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	count := 0
	for k, u := range sm.data {
		if u != nil {
			proc(k, u)
			count += 1
		}
	}
	return count
}

func (sm *SafeMap) List() <-chan interface{} {
	ret := make(chan interface{}, 1024)

	go func() {
		sm.lock.RLock()
		defer sm.lock.RUnlock()

		for _, u := range sm.data {
			ret <- u
		}
		ret <- nil //Finish channel by nil.
	}()

	return ret
}

type SafeStrMap struct {
	size int
	data map[string]interface{}
	lock sync.RWMutex
}

func NewSafeStrMap(size int) *SafeStrMap {
	calSize := size
	if calSize == 0 {
		calSize = 128
	}
	return &SafeStrMap{
		size: size,
		data: make(map[string]interface{}, calSize),
	}
}

func (sm *SafeStrMap) Len() int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return len(sm.data)
}

func (sm *SafeStrMap) Set(k string, v interface{}) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if sm.size != 0 && len(sm.data) >= sm.size {
		return NewErr("SageMap.Set already full")
	}
	sm.data[k] = v
	return nil
}

func (sm *SafeStrMap) Del(k string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if _, ok := sm.data[k]; ok {
		delete(sm.data, k)
	}
}

func (sm *SafeStrMap) Get(k string) interface{} {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return sm.data[k]
}

func (sm *SafeStrMap) GetEx(k string) (interface{}, bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	v, ok := sm.data[k]
	return v, ok
}

func (sm *SafeStrMap) Iter(proc func(k string, v interface{})) int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	count := 0
	for k, u := range sm.data {
		if u != nil {
			proc(k, u)
			count += 1
		}
	}
	return count
}

func (sm *SafeStrMap) List() <-chan interface{} {
	ret := make(chan interface{}, 1024)

	go func() {
		sm.lock.RLock()
		defer sm.lock.RUnlock()

		for _, u := range sm.data {
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
	data interface{}
	lock sync.RWMutex
}

func NewSafeVar() *SafeVar {
	return &SafeVar{}
}

func (sv *SafeVar) Set(v interface{}) {
	sv.lock.Lock()
	defer sv.lock.Unlock()
	sv.data = v
}

func (sv *SafeVar) Get() interface{} {
	sv.lock.RLock()
	defer sv.lock.RUnlock()
	return sv.data
}

func (sv *SafeVar) GetWithFunc(proc func(v interface{})) {
	sv.lock.RLock()
	defer sv.lock.RUnlock()
	proc(sv.data)
}
