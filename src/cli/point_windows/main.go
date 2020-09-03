// +build windows

package main

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olp"
)

func main() {
	c := config.NewPoint()
	p := olp.NewPoint(c)
	p.Initialize()
	p.Start()
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
