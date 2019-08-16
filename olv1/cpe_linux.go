package main

import (
    "log"
    "flag"
    "fmt"
    "os/signal"
    "syscall"
    "time"
    "os"
    
    "github.com/songgao/water"
    "github.com/milosgajdos83/tenus"
    "github.com/danieldin95/openlan-go/olv1/olv1"
    "github.com/danieldin95/openlan-go/olv1/cpe"
)

type Cpe struct {
    verbose int
    tcpwroker *olv1cpe.TcpWroker 
    tapwroker *olv1cpe.TapWroker
}

func NewCpe(client *olv1.TcpClient, ifce *water.Interface, 
            name string, password string, ifmtu int, verbose int) (this *Cpe){
    this = &Cpe {
        verbose: verbose,
        tapwroker : olv1cpe.NewTapWoker(ifce, ifmtu, verbose),
        tcpwroker : olv1cpe.NewTcpWoker(client, name, password, ifmtu, verbose),
    }

    return 
}

func (this *Cpe) Start() {
    go this.tapwroker.GoRecv(this.tcpwroker.DoSend)
    go this.tapwroker.GoLoop()

    go this.tcpwroker.GoRecv(this.tapwroker.DoSend)
    go this.tcpwroker.GoLoop()
}

func NewIfce(devtype water.DeviceType) (ifce *water.Interface) {
    ifce, err := water.New(water.Config {
        DeviceType: devtype,
    })
    if err != nil {
        log.Fatal(err)
    }

    return 
}

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
    addr := flag.String("addr", "openlan.net:10002",  "the server connect to")
    verbose := flag.Int("verbose", 0x00, "open verbose")
    ifmtu := flag.Int("ifmtu", 1514, "the interface MTU include ethernet")
    name := flag.String("name", "openlan",  "the name login to")
    password := flag.String("password", "openlan.net",  "the password login to")

    flag.Parse()

    ifce := NewIfce(water.TAP)
    UpLink(ifce.Name())

    client := olv1.NewTcpClient(*addr, *verbose)
    cpe := NewCpe(client, ifce, *name, *password, *ifmtu, *verbose)

    if err := client.Connect(); err != nil {
        log.Printf("main %s\n", err)
    }

    cpe.Start()

    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        client.Close()
        ifce.Close()
        fmt.Println("Done!")
        os.Exit(0)
    }()

    fmt.Println("Please enter CTRL+C to exit...")
    for {
        time.Sleep(1000 * time.Second)
    }
}
