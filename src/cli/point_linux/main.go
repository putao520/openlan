// +build linux

package main

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olap"
)

func main() {
	c := config.NewPoint()
	p := olap.NewPoint(c)
	libol.PreNotify()
	p.Initialize()
	libol.Go(p.Start)
	libol.SdNotify()
	libol.Wait()
	p.Stop()
}
