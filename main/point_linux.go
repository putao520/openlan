package main

import (
	"fmt"
	"github.com/danieldin95/openlan-go/config"
	"github.com/danieldin95/openlan-go/point"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := config.NewPoint()
	s := config.NewScript(c.Script)

	s.CallBefore()
	p := point.NewPoint(c)
	p.Start()
	s.CallAfter(p.IfName(), p.IfAddr())

	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGKILL)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C
	signal.Notify(x, os.Interrupt, syscall.SIGABRT) //Exit

	<-x
	s.CallExit()
	p.Stop()
	fmt.Println("Done!")
}
