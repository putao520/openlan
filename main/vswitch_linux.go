package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lightstar-dev/openlan-go/vswitch"
)

type VSwitch struct {
	Wroker *vswitch.VSwitchWroker
}

func NewVSwitch(c *vswitch.Config) (this *VSwitch) {
	server := vswitch.NewTcpServer(c)
	this = &VSwitch{
		Wroker: vswitch.NewVSwitchWroker(server, c),
	}
	return
}

func GoHttp(ope *VSwitch, c *vswitch.Config) {
	http := vswitch.NewVSwitchHttp(ope.Wroker, c)
	http.GoStart()
}

func main() {
	c := vswitch.NewConfig()
	vs := NewVSwitch(c)
	vs.Wroker.Start()

	go GoHttp(vs, c)

	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-x
		vs.Wroker.Close()
		fmt.Println("Done!")
		os.Exit(0)
	}()

	for {
		time.Sleep(1000 * time.Second)
	}
}
