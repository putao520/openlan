package olsw

import (
	"fmt"
	"github.com/danieldin95/openlan/pkg/olsw/cache"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSwitch_LoadPass(t *testing.T) {
	sw := &Switch{}
	sw.LoadPass("../../.password.no")
	sw.LoadPass("../../packaging/resource/password.example")
	for user := range cache.User.List() {
		if user == nil {
			break
		}
		fmt.Printf("%v\n", user)
	}
	assert.Equal(t, 2, cache.User.Users.Len(), "notEqual")
}
