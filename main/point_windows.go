package main

import (
	"fmt"
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/point"
)

func main() {
	c := config.NewPoint()
	s := config.NewScript(c.Script)

	s.CallBefore()
	p := point.NewPoint(c)
	p.Start()
	s.CallAfter(p.IfName(), p.IfAddr)

	for {
		fmt.Println("Please press enter `q` to exit...")

		input := ""
		if fmt.Scanln(&input); input == "q" {
			break
		}
	}

	s.CallExit()
	p.Stop()
	fmt.Println("Done!")
}
