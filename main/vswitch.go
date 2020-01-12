package main

import (
	"fmt"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/vswitch"
	"os"
	"os/signal"
	"syscall"
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
	signal.Notify(x, os.Interrupt, syscall.SIGKILL)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT) //CTL+C

	<-x
	s.CallExit()
	vs.Stop()
	fmt.Println("Done!")
}
