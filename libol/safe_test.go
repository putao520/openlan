package libol

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestSafeMap(t *testing.T) {
	m := NewSafeMap(1024)
	m.Set("hi", 1)
	i := m.Get("hi")
	assert.Equal(t, i, 1, "be the same.")

	a := 3
	m.Set("hip", &a)
	c := m.Get("hip").(*int)
	assert.Equal(t, c, &a, "be the same.")
	assert.Equal(t, 2, m.Len(), "be the same.")

	for i := 0; i < 1024; i++ {
		m.Set(i, i)
	}
	assert.Equal(t, 1024, m.Len(), "")
	fmt.Printf("TestSafeMap.size: %d\n", m.Len())
	for i := 0; i < 1024; i++ {
		m.Del(i)
	}
	assert.Equal(t, 2, m.Len(), "")

	m.Del("hi")
	ii := m.Get("hi")
	assert.Equal(t, ii, nil, "be the same.")
	assert.Equal(t, 1, m.Len(), "be the same.")

	iii := m.Get("hello")
	assert.Equal(t, iii, nil, "be the same.")

	var ret interface{}
	if ret != nil {
		ap := ret.(*int)
		assert.Equal(t, nil, ap, "be the same")
	}
}

func TestZeroSafeMap(t *testing.T) {
	m := NewSafeMap(0)
	m.Set("hi", 1)
	i := m.Get("hi")
	assert.Equal(t, i, 1, "be the same.")

	a := 3
	m.Set("hip", &a)
	c := m.Get("hip").(*int)
	assert.Equal(t, c, &a, "be the same.")
	assert.Equal(t, 2, m.Len(), "be the same.")

	for i := 0; i < 1024; i++ {
		m.Set(i, i)
	}
	assert.Equal(t, 1026, m.Len(), "")
	fmt.Printf("TestZeroSafeMap.size: %d\n", m.Len())
	for i := 0; i < 1024; i++ {
		m.Del(i)
	}
	assert.Equal(t, 2, m.Len(), "")
	m.Del("hi")
	ii := m.Get("hi")
	assert.Equal(t, ii, nil, "be the same.")
	assert.Equal(t, 1, m.Len(), "be the same.")

	iii := m.Get("hello")
	assert.Equal(t, iii, nil, "be the same.")
}

func TestZeroSafeMapIter(t *testing.T) {
	m := NewSafeMap(0)
	c := 0
	for i := 0; i < 1024; i++ {
		c += i
		m.Set(i, i)
	}
	ct := 0
	m.Iter(func(k interface{}, v interface{}) {
		ct += v.(int)
	})
	assert.Equal(t, ct, c, "be the same")

	ms := NewSafeStrMap(0)
	cm := 0
	for i := 1024; i < 1024+1024; i++ {
		cm += i
		ms.Set(fmt.Sprintf("%d", i), i)
	}
	cmt := 0
	ms.Iter(func(k string, v interface{}) {
		cmt += v.(int)
	})
	assert.Equal(t, cmt, cm, "be the same")
}

func TestSafeVar(t *testing.T) {
	v := NewSafeVar()
	a := 3
	c := 0
	v.Set(2)
	v.GetWithFunc(func(v interface{}) {
		c = a + v.(int)
	})
	assert.Equal(t, 5, c, "")
}

func BenchmarkMapGet(b *testing.B) {
	m := make(map[string]int, 2)
	m["hi"] = 2

	for i := 0; i < b.N; i++ {
		v := m["hi"]
		assert.Equal(b, v, 2, "")
	}
}

func BenchmarkMapGetWithLock(b *testing.B) {
	m := make(map[string]int, 2)
	m["hi"] = 2
	lock := sync.RWMutex{}

	for i := 0; i < b.N; i++ {
		lock.RLock()
		v := m["hi"]
		lock.RUnlock()
		assert.Equal(b, v, 2, "")
	}
}

func BenchmarkSafeMapGet(b *testing.B) {
	m := NewSafeMap(2)
	m.Set("hi", 2)

	for i := 0; i < b.N; i++ {
		v := m.Get("hi").(int)
		assert.Equal(b, v, 2, "")
	}
}

func BenchmarkSafeStrMapGet(b *testing.B) {
	m := NewSafeStrMap(2)
	m.Set("hi", 2)

	for i := 0; i < b.N; i++ {
		v := m.Get("hi").(int)
		assert.Equal(b, v, 2, "")
	}
}
