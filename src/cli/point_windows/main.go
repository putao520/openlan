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
				fmt.Printf("%-15s: %s\n", "UUID", p.UUID())
				fmt.Printf("%-15s: %d\n", "UpTime", p.UpTime())
				fmt.Printf("%-15s: %s\n", "Device", p.IfName())
				fmt.Printf("%-15s: %v\n", "Record", p.Record())
				if client != nil {
					fmt.Printf("%-15s: %s\n", "Status", client.Status())
					fmt.Printf("%-15s: %v\n", "Statistics", client.Statistics())
				}
			}
		}
	})
	libol.Wait()
	p.Stop()
}
