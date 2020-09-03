// +build darwin

package main

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olp"
)

func main() {
	c := config.NewPoint()
	p := olp.NewPoint(c)
	p.Initialize()
	p.Start()
	libol.Wait()
	p.Stop()
}
