package main 

import (
    "fmt"
    "os/signal"
    "syscall"
    "time"
    "os"
    "log"

    "github.com/danieldin95/openlan-go/olv1/ope"
)

type Ope struct {
    Wroker *olv1ope.OpeWroker
}

func NewOpe(c *olv1ope.Config) (this *Ope){
    server := olv1ope.NewTcpServer(c)
    this = &Ope {
        Wroker: olv1ope.NewOpeWroker(server, c),
    }
    return 
}

func GoHttp(ope *Ope, c *olv1ope.Config) {
    http := olv1ope.NewOpeHttp(ope.Wroker, c)
    http.GoStart()
}

func main() {
    c := olv1ope.NewConfig()
    log.Printf("Debug| main.config: %s", c)
    ope := NewOpe(c)
    ope.Wroker.Start()

    go GoHttp(ope, c)
    
    x := make(chan os.Signal)
    signal.Notify(x, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- x
        ope.Wroker.Close()
        fmt.Println("Done!")
        os.Exit(0)
    }()

    fmt.Println("Please enter CTRL+C to exit...")
    for {
        time.Sleep(1000 * time.Second)
    }
}