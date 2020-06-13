// +build darwin

package main

import (
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/point"
)

func main() {
	c := config.NewPoint()
	p := point.NewPoint(c)
	p.Initialize()
	p.Start()
	libol.Wait()
	p.Stop()
}
