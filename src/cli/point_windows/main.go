// +build windows

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
	if c.Terminal == "on" {
		olap.NewTerminal(p).Start()
	} else {
		libol.Wait()
	}
	p.Stop()
}
