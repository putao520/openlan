package main

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olsw"
	"github.com/danieldin95/openlan-go/src/olsw/storage"
)

func main() {
	c := config.NewSwitch()
	storage.Init(c.Perf)
	s := olsw.NewSwitch(c)
	libol.PreNotify()
	s.Initialize()
	s.Start()
	libol.SdNotify()
	libol.Wait()
	s.Stop()
}
