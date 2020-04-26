// +build darwin

package main

import (
	"fmt"
	"github.com/danieldin95/openlan-go/main/config"
	"github.com/danieldin95/openlan-go/point"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := config.NewPoint()
	p := point.NewPoint(c)
	p.Start()

	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGKILL)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C

	<-x
	p.Stop()
	fmt.Println("Done!")
}
