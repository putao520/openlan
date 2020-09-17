// +build darwin

package main

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olap"
)

func main() {
	c := config.NewPoint()
	p := olap.NewPoint(c)
	p.Initialize()
	libol.Go(p.Start)
	if !c.Daemon {
		t := olap.NewTerminal(p)
		libol.Go(t.Start)
	}
	libol.Wait()
	p.Stop()
}
