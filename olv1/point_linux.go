package main

import (
    "log"
    "fmt"
    "os/signal"
    "syscall"
    "time"
    "os"

    "github.com/milosgajdos83/tenus"
    "github.com/danieldin95/openlan-go/olv1/cpe"
)

func UpLink(name string) error {
    link, err := tenus.NewLinkFrom(name)
    if err != nil {
        log.Printf("Error|main.UpLink: Get ifce %s: %s", name, err)
        return err
    }
    
    if err := link.SetLinkUp(); err != nil {
        log.Printf("Error|main.UpLink: %s : %s", name, err)
        return err
    }
    
    return nil
}

func main() {
    c := point.NewConfig()
    log.Printf("Debug| main.config: %s", c)

    cpe := point.NewCpe(c)

    UpLink(cpe.Ifce.Name())
    cpe.Start()

    x := make(chan os.Signal)
    signal.Notify(x, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- x
        cpe.Close()
        fmt.Println("Done!")
        os.Exit(0)
    }()

    fmt.Println("Please enter CTRL+C to exit...")
    for {
        time.Sleep(1000 * time.Second)
    }
}
