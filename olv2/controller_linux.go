package main 

import (
	"log"
    "fmt"
    "os/signal"
    "syscall"
    "time"
    "os"
    
	"github.com/danieldin95/openlan-go/olv2/controller"
)

func main() {
	c := controller.NewConfig()
    log.Printf("Debug| main.config: %s", c)

    ctl := controller.NewController(c)

    ctl.Start()

    x := make(chan os.Signal)
    signal.Notify(x, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- x
        ctl.Stop()
        fmt.Println("Done!")
        os.Exit(0)
    }()

    fmt.Println("Please enter CTRL+C to exit...")
    for {
        time.Sleep(1000 * time.Second)
    }

}