package ctrls

import (
	"github.com/danieldin95/openlan/pkg/olctl/libctrl"
)

type OnLine struct {
	libctrl.Listen
	cc *CtrlC
}
