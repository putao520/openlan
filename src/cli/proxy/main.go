package main

import (
	"github.com/danieldin95/openlan/src/config"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/proxy"
)

func main() {
	c := config.NewProxy()
	libol.SetLogger(c.Log.File, c.Log.Verbose)

	p := proxy.NewProxy(c)
	libol.PreNotify()
	p.Initialize()
	libol.Go(p.Start)
	libol.SdNotify()
	libol.Wait()
	p.Stop()
}
