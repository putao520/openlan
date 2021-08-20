package olsw

import (
	"fmt"
	"github.com/danieldin95/openlan/src/olsw/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSwitch_LoadPass(t *testing.T) {
	sw := &Switch{}
	sw.LoadPass("../../.password.no")
	sw.LoadPass("../../packaging/resource/password.example")
	for user := range store.User.List() {
		if user == nil {
			break
		}
		fmt.Printf("%v\n", user)
	}
	assert.Equal(t, 2, store.User.Users.Len(), "notEqual")
}
