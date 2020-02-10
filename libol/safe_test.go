package libol

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSafeMap(t *testing.T) {
	m := NewSafeMap(1024)
	m.Set("hi", 1)
	i := m.Get("hi")
	assert.Equal(t, i, 1, "The two words should be the same.")

	a :=3
	m.Set("hip", &a)
	c := m.Get("hip").(*int)
	assert.Equal(t, c, &a, "The two words should be the same.")
	assert.Equal(t, 2, m.Len(), "The two words should be the same.")

	m.Del("hi")
	ii := m.Get("hi")
	assert.Equal(t, ii, nil, "The two words should be the same.")
	assert.Equal(t, 1, m.Len(), "The two words should be the same.")

	iii := m.Get("hello")
	assert.Equal(t, iii, nil, "The two words should be the same.")
}
