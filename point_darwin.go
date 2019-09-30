package main

import (
    "fmt"
    "os/signal"
    "syscall"
    "time"
    "os"

    "github.com/lightstar-dev/openlan-go/libol"
    "github.com/lightstar-dev/openlan-go/point"
)

func main() {
    c := point.NewConfig()
    if c.Verbose != 0 {
        libol.Debug("main.config: %s", c)
    }

    p := point.NewPoint(c)
    p.Start()

    x := make(chan os.Signal)
    signal.Notify(x, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- x
        p.Close()
        fmt.Println("Done!")
        os.Exit(0)
    }()

    for {
        time.Sleep(1000 * time.Second)
    }
}
