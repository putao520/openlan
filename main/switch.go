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
	go s.Start()

	libol.Wait()
	_ = s.Stop()
	libol.SdNotify()
}
