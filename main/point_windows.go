package main

import (
	"fmt"
	"github.com/lightstar-dev/openlan-go/config"
	"github.com/lightstar-dev/openlan-go/point"
)

func main() {
	c := config.NewPoint()
	p := point.NewPoint(c)
	p.Start()

	for {
		fmt.Println("Please press enter `q` to exit...")

		input := ""
		if fmt.Scanln(&input); input == "q" {
			break
		}
	}

	p.Stop()
	fmt.Println("Done!")
}
