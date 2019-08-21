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
    Wroker *vswitch.OpeWroker
}

func NewOpe(c *vswitch.Config) (this *Ope){
    server := vswitch.NewTcpServer(c)
    this = &Ope {
        Wroker: vswitch.NewOpeWroker(server, c),
    }
    return 
}

func GoHttp(ope *Ope, c *vswitch.Config) {
    http := vswitch.NewOpeHttp(ope.Wroker, c)
    http.GoStart()
}

func main() {
    c := vswitch.NewConfig()
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