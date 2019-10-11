package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lightstar-dev/openlan-go/vswitch"
)

func main() {
	c := vswitch.NewConfig()
	vs := vswitch.NewVSwitch(c)
	vs.Start()

	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-x
		vs.Stop()
		fmt.Println("Done!")
		os.Exit(0)
	}()

	for {
		time.Sleep(1000 * time.Second)
	}
}
