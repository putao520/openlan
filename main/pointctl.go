package main

import (
	"github.com/lightstar-dev/openlan-go/point"
	"github.com/lightstar-dev/openlan-go/point/models"
)

func main() {
	c := models.NewConfig()
	p := point.NewCommand(c)
	p.Start()
	p.Loop()
}
