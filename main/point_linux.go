package main

import (
	"fmt"
	"github.com/danieldin95/openlan-go/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danieldin95/openlan-go/point"
)

func main() {
	c := config.NewPoint()
	s := config.NewScript(c.Script)

	s.CallBefore()
	p := point.NewPoint(c)
	p.Start()
	s.CallAfter()

	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-x
		s.CallExit()
		p.Stop()
		fmt.Println("Done!")
		os.Exit(0)
	}()

	for {
		time.Sleep(1000 * time.Second)
	}
}
