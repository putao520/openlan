package ctrls

import (
	"github.com/danieldin95/openlan/src/olctl/libctrl"
)

type OnLine struct {
	libctrl.Listen
	cc *CtrlC
}
