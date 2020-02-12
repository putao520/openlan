package libol

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrettyTime(t *testing.T) {
	var s string

	s = PrettyTime(59)
	assert.Equal(t, "0m59s", s, "be the same.")

	s = PrettyTime(60*2 + 8)
	assert.Equal(t, "2m8s", s, "be the same.")

	s = PrettyTime(3600 + 1)
	assert.Equal(t, "1h0m", s, "be the same.")

	s = PrettyTime(3600 + 61)
	assert.Equal(t, "1h1m", s, "be the same.")

	s = PrettyTime(3600 + 60*59)
	assert.Equal(t, "1h59m", s, "be the same.")

	s = PrettyTime(86400)
	assert.Equal(t, "1d0h", s, "be the same.")

	s = PrettyTime(86400 + 3600*5 + 59)
	assert.Equal(t, "1d5h", s, "be the same.")

	s = PrettyTime(86400 + 3600*23 + 59)
	assert.Equal(t, "1d23h", s, "be the same.")
}
