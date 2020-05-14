package main

import (
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/vswitch"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := config.NewVSwitch()
	vs := vswitch.NewVSwitch(c)

	libol.PreNotify()
	_ = vs.Start()
	libol.SdNotify()
	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGKILL)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C

	<-x
	_ = vs.Stop()
	fmt.Println("Done!")
}
