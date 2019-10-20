package main

import (
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/point"
)

func main() {
	c := config.NewPoint()
	p := point.NewCommand(c)
	p.Start()
	p.Loop()
}
