// +build linux

package main

import (
	"github.com/danieldin95/openlan/src/config"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/olap"
)

func main() {
	c := config.NewPoint()
	p := olap.NewPoint(c)
	// terminal off for linux service, on for open a terminal
	// and others just wait.
	if c.Terminal == "off" {
		libol.PreNotify()
	}
	p.Initialize()
	libol.Go(p.Start)
	if c.Terminal == "on" {
		t := olap.NewTerminal(p)
		t.Start()
	} else if c.Terminal == "off" {
		libol.SdNotify()
		libol.Wait()
	} else {
		libol.Wait()
	}
	p.Stop()
}
