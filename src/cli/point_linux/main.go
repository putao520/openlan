// +build linux

package main

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olp"
)

func main() {
	c := config.NewPoint()
	p := olp.NewPoint(c)
	libol.PreNotify()
	p.Initialize()
	p.Start()
	libol.SdNotify()
	libol.Wait()
	p.Stop()
}
