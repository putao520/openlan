package main

import (
	"github.com/lightstar-dev/openlan-go/point"
)

func main() {
	c := point.NewConfig()
	p := point.NewCommand(c)
	p.Start()
	p.Loop()
}
