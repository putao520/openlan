package ctrls

import (
	"github.com/danieldin95/openlan-go/controller/libctrl"
)

type OnLine struct {
	libctrl.Listen
	cc *CtrlC
}
