package main 


import (
    "fmt"
    "flag"
    "net/http"
    "html"
    "log"

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

func NewHttp(ope *Ope, listen string) {
    http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
    })

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        body := "remote address, device name\n"
        for client, ifce := range ope.Wroker.Clients {
            body += fmt.Sprintf("%s, %s\n", client.GetAddr(), ifce.Name())
        }
        fmt.Fprintf(w, body)
    })

    log.Fatal(http.ListenAndServe(listen, nil))
}

func main() {
    br := flag.String("br", "",  "the bridge name")
    http := flag.String("http", "0.0.0.0:10082",  "the http listen on")
    addr := flag.String("addr", "0.0.0.0:10002",  "the server listen on")
    verbose := flag.Int("verbose", 0x00, "open verbose")
    ifmtu := flag.Int("ifmtu", 1514, "the interface MTU include ethernet")

    flag.Parse()

    ope := NewOpe(*addr, *ifmtu, *br, *verbose)
    ope.Wroker.Start()

    go NewHttp(ope, *http)
    
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