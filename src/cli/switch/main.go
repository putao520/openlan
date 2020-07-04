package main

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/switch"
)

func main() {
	c := config.NewSwitch()
	s := _switch.NewSwitch(c)
	if c.Prof != "" {
		f := libol.Prof{File: c.Prof}
		f.Start()
		defer f.Stop()
	}
	libol.PreNotify()
	s.Initialize()
	s.Start()
	libol.SdNotify()
	libol.Wait()
	s.Stop()
}
