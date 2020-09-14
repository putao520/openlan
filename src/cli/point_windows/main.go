// +build windows

package main

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/olap"
)

func main() {
	c := config.NewPoint()
	p := olap.NewPoint(c)
	p.Initialize()
	libol.Go(p.Start)
	libol.Go(func() {
		for {
			input := ""
			if _, _ = fmt.Scanln(&input); input == "quit" {
				fmt.Printf("press `CTRL+C` to exit...\n")
				break
			}
			if input == "s" || input == "" {
				client := p.Client()
				fmt.Printf("UUID  : %s\n", p.UUID())
				if client != nil {
					fmt.Printf("Status : %s\n", client.Status())
				}
				fmt.Printf("Uptime: %d\n", p.UpTime())
				fmt.Printf("Device: %s\n", p.IfName())
				if client != nil {
					fmt.Printf("Statistics: %v\n", client.Statistics())
				}
			}
		}
	})
	libol.Wait()
	p.Stop()
}
