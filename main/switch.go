package main

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/switch"
)

func main() {
	c := config.NewSwitch()
	s := _switch.NewSwitch(c)

	libol.PreNotify()
	s.Initialize()
	_ = s.Start()
	libol.SdNotify()
	libol.Wait()
	_ = s.Stop()
}
