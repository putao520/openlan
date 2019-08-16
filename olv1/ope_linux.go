package main 

import (
    "fmt"
    "flag"
    "os/signal"
    "syscall"
    "time"
    "os"

    "github.com/danieldin95/openlan-go/olv1/ope"
)

type Ope struct {
    Wroker *olv1ope.OpeWroker
}

func NewOpe(addr string, ifmtu int, brname string, verbose int) (this *Ope){
    server := olv1ope.NewTcpServer(addr, verbose)
    this = &Ope {
        Wroker: olv1ope.NewOpeWroker(server, brname, verbose),
    }
    return 
}

func GoHttp(ope *Ope, listen string, token string) {
    http := olv1ope.NewOpeHttp(ope.Wroker, listen, token)
    http.GoStart()
}

func main() {
    br := flag.String("br", "",  "the bridge name")
    verbose := flag.Int("verbose", 0x00, "open verbose")
    http := flag.String("http", "0.0.0.0:10082",  "the http listen on")
    addr := flag.String("addr", "0.0.0.0:10002",  "the server listen on")
    ifmtu := flag.Int("ifmtu", 1514, "the interface MTU include ethernet")
    token := flag.String("token", "dontUseDefault", "Administrator token")

    flag.Parse()

    ope := NewOpe(*addr, *ifmtu, *br, *verbose)
    ope.Wroker.Start()

    go GoHttp(ope, *http, *token)
    
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        ope.Wroker.Close()
        fmt.Println("Done!")
        os.Exit(0)
    }()

    fmt.Println("Please enter CTRL+C to exit...")
    for {
        time.Sleep(1000 * time.Second)
    }
}