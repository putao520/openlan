// +build windows

package main

import (
	"github.com/danieldin95/openlan/pkg/config"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/danieldin95/openlan/pkg/olap"
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
