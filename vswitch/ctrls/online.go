package ctrls

import (
	"github.com/danieldin95/openlan-go/controller/ctl"
)

type OnLine struct {
	ctl.Listen
	cc *CtrlC
}
