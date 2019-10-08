package libol

import "sync"

//m := NewSMap(1024)
//m.Set("hi", 1)
//a=3
//m.Set("hip", &a)
//c := m.Get("hip").(*int)
//fmt.Printf("%s\n%d\n", m, *c)

type SMap struct {
	Data map[interface{}]interface{}
	Lock sync.RWMutex
}

func NewSMap(size int) *SMap {
	return &SMap{
		Data: make(map[interface{}]interface{}, size),
	}
}

func (sm *SMap) Set(k interface{}, v interface{}) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	sm.Data[k] = v
}

func (sm *SMap) Del(k interface{})  {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()

	if _, ok := sm.Data[k]; ok {
		delete(sm.Data, k)
	}
}

func (sm *SMap) Get(k interface{}) interface{} {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	return sm.Data[k]
}

func (sm *SMap) GetEx(k string) (interface{}, bool) {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	v, ok := sm.Data[k]
	return v, ok
}

// a := SVar
// a.Set(0x01)
// a.Get().(int)

type SVar struct {
	Data interface{}
	Lock sync.RWMutex
}

func NewSVar(size int) *SVar {
	return &SVar{}
}

func (sv *SVar) Set(v interface{}) {
	sv.Lock.Lock()
	defer sv.Lock.Unlock()
	sv.Data = v
}

func (sv *SVar) Get() interface{} {
	sv.Lock.RLock()
	defer sv.Lock.RUnlock()
	return sv.Data
}


