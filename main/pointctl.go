package main

import (
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/point"
)

func main() {
	c := config.NewPoint()
	p := point.NewCommand(c)
	p.Start()
	p.Loop()
}
