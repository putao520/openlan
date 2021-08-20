package main

import (
	"github.com/danieldin95/openlan/src/config"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/olsw"
	"github.com/danieldin95/openlan/src/olsw/store"
)

func main() {
	c := config.NewSwitch()
	libol.SetLogger(c.Log.File, c.Log.Verbose)
	libol.Debug("main %s", c)
	store.Init(&c.Perf)
	s := olsw.NewSwitch(c)
	libol.PreNotify()
	s.Initialize()
	s.Start()
	libol.SdNotify()
	libol.Wait()
	s.Stop()
}
