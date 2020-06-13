package libol

type gos struct {
	lock  Locker
	total uint64
	current map[interface{}]interface{}
}

var Gos = gos{
	current: make(map[interface{}]interface{}, 32),
}

func (t *gos) Add(call interface{}, obj interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.total++
	t.current[obj] = call
	Debug("gos.Add %d %p %v", t.total, obj, call)
}

func (t *gos) Del(obj interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if _, ok := t.current[obj]; ok {
		t.total--
		delete(t.current, obj)
		Debug("gos.Del %d %p", t.total, obj)
	}
}

func Go(call func(), obj interface{}) {
	go func() {
		Gos.Add(call, obj)
		call()
		Gos.Del(obj)
	}()
}
