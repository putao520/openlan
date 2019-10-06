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
	Wroker *vswitch.Worker
}

func NewVSwitch(c *vswitch.Config) (this *VSwitch) {
	server := vswitch.NewTcpServer(c)
	this = &VSwitch{
		Wroker: vswitch.NewWorker(server, c),
	}
	return
}

func GoHttp(ope *VSwitch, c *vswitch.Config) {
	http := vswitch.NewHttp(ope.Wroker, c)
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
		vs.Wroker.Stop()
		fmt.Println("Done!")
		os.Exit(0)
	}()

	for {
		time.Sleep(1000 * time.Second)
	}
}
