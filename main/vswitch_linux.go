package main

import (
	"fmt"
	"github.com/lightstar-dev/openlan-go/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lightstar-dev/openlan-go/vswitch"
)

func main() {
	c := config.NewVSwitch()
	s := config.NewScript(c.Script)

	s.CallBefore()
	vs := vswitch.NewVSwitch(c)
	vs.Start()
	s.CallAfter()

	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-x
		s.CallExit()
		vs.Stop()
		fmt.Println("Done!")
		os.Exit(0)
	}()

	for {
		time.Sleep(1000 * time.Second)
	}
}
