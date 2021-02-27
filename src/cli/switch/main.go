package main

import (
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olsw"
	"github.com/danieldin95/openlan-go/src/olsw/store"
)

func main() {
	c := config.NewSwitch()
	libol.SetLogger(c.Log.File, c.Log.Verbose)
	store.Init(c.Perf)
	s := olsw.NewSwitch(c)
	libol.PreNotify()
	s.Initialize()
	s.Start()
	libol.SdNotify()
	libol.Wait()
	s.Stop()
}
