package main 


import (
    "fmt"
    "flag"

    "github.com/danieldin95/openlan-go/olv1"
)

type Ope struct {
    Wroker *olv1.OpeWroker
}

func NewOpe(addr string, ifmtu int, verbose int) (this *Ope){
    server := olv1.NewTcpServer(addr, verbose)
    this = &Ope {
        Wroker: olv1.NewOpeWroker(server, "", verbose),
    }
    return 
}

func main() {
    addr := flag.String("addr", "0.0.0.0:10001",  "the server address")
    verbose := flag.Int("verbose", 0x00, "open verbose")
    ifmtu := flag.Int("ifmtu", 1514, "the interface MTU include ethernet")

    flag.Parse()

    ope := NewOpe(*addr, *ifmtu, *verbose)
    ope.Wroker.Start()

    for {
        var input string

        fmt.Println("Please press enter `q` to exit...")
        if fmt.Scanln(&input); input == "q" {
            break
        }
    }
    
	ope.Wroker.Close()
	fmt.Println("Done!")
}