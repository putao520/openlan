package main 

import (
    "fmt"
    "os/signal"
    "syscall"
    "time"
    "os"
    "log"

    "github.com/danieldin95/openlan-go/vswitch"
)

type VSwitch struct {
    Wroker *vswitch.VSwitchWroker
}

func NewVSwitch(c *vswitch.Config) (this *VSwitch){
    server := vswitch.NewTcpServer(c)
    this = &VSwitch {
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
    log.Printf("Debug| main.config: %s", c)
    vs := NewVSwitch(c)
    vs.Wroker.Start()

    go GoHttp(vs, c)
    
    x := make(chan os.Signal)
    signal.Notify(x, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- x
        vs.Wroker.Close()
        fmt.Println("Done!")
        os.Exit(0)
    }()

    fmt.Println("Please enter CTRL+C to exit...")
    for {
        time.Sleep(1000 * time.Second)
    }
}
