// +build darwin

package main

import (
	"github.com/danieldin95/openlan/src/config"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/olap"
)

func main() {
	c := config.NewPoint()
	p := olap.NewPoint(c)
	p.Initialize()
	libol.Go(p.Start)
	if c.Terminal == "on" {
		t := olap.NewTerminal(p)
		t.Start()
	} else {
		libol.Wait()
	}
	p.Stop()
}
