package main

import (
	"fmt"
	"github.com/lightstar-dev/openlan-go/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lightstar-dev/openlan-go/point"
)

func main() {
	c := config.NewPoint()
	p := point.NewPoint(c)

	p.Start()

	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-x
		p.Stop()
		fmt.Println("Done!")
		os.Exit(0)
	}()

	for {
		time.Sleep(1000 * time.Second)
	}
}
