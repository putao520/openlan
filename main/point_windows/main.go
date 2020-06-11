// +build windows

package main

import (
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/point"
)

func main() {
	c := config.NewPoint()
	p := point.NewPoint(c)
	go p.Start()

	go func() {
		for {
			input := ""
			if _, _ = fmt.Scanln(&input); input == "quit" {
				fmt.Printf("press `CTRL+C` to exit...\n")
				break
			}
			if input == "s" || input == "state" {
				fmt.Printf("UUID  : %s\n", p.UUID())
				fmt.Printf("State : %s\n", p.State())
				fmt.Printf("Uptime: %d\n", p.UpTime())
				fmt.Printf("Device: %s\n", p.IfName())
			}
		}
	}()
	libol.Wait()
	p.Stop()
}
