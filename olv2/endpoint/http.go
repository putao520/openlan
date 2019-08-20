package endpoint

import (
    "os"
    "fmt"
    "html"
    "log"
    "net/http"

    "github.com/danieldin95/openlan-go/olv2/openlanv2"
)

type Http struct {
    Point *Endpoint
    Token string
    TokenFile string
    Listen string
    //
    verbose bool
}

func NewHttp(point *Endpoint, c *Config)(this *Http) {
    token := c.Token
    if token == "" {
        token = openlanv2.GenUUID(16)
    }

    this = &Http {
        Point: point,
        Listen: c.HttpListen,
        Token: token,
        TokenFile: c.TokenFile,
        verbose: c.Verbose,
    }

    this.SaveToken()
    http.HandleFunc("/", this.Index)
    http.HandleFunc("/hi", this.Hi)
    http.HandleFunc("/mac", this.Mac)

    return 
}

func (this *Http) SaveToken() error {
    log.Printf("Info| Http.SaveToken: Token: %s", this.Token)

    f, err := os.OpenFile(this.TokenFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
    defer f.Close()
    if err != nil {
        log.Printf("Error| Http.SaveToken: %s", err)
        return err
    }

    if _, err := f.Write([]byte(this.Token)); err != nil {
        log.Printf("Error| Http.SaveToken: %s", err)
        return err
    }

    return nil
}

func (this *Http) GoStart() error {
    log.Printf("Info| Http.GoStart on %s", this.Listen)
    if err := http.ListenAndServe(this.Listen, nil); err != nil {
        log.Printf("Error| Http.GoStart on %s: %s", this.Listen, err)
        return err
    }
    return nil
}

func (this *Http) IsAuth(w http.ResponseWriter, r *http.Request) bool {
    token, pass, ok := r.BasicAuth()
    if this.verbose {
        log.Printf("token: %s, pass: %s", token, pass)
    }
    if !ok  || token != this.Token {
        w.Header().Set("WWW-Authenticate", "Basic")
        http.Error(w, "Authorization Required.", 401)
        return false
    }

    return true
}

func (this *Http) Hi(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi %s %q", r.Method, html.EscapeString(r.URL.Path))

    for name, headers := range r.Header {
        for _, h := range headers {
            log.Printf("Info| Http.Hi %v: %v", name, h)
        }
    }
}

func (this *Http) Index(w http.ResponseWriter, r *http.Request) {
    if (!this.IsAuth(w, r)) {
        return
    }

    switch (r.Method) {
    case "GET":  
        body := "uptime,update,uuid,remoteaddr,receipt,transmission,error\n"
        for peer := range this.Point.GetPeers() {
            if peer == nil {
                break
            }
            body += fmt.Sprintf("%d, %d, %s, %s, %d, %d, %d\n", 
                                peer.UpTime(), peer.UpdateTime(), peer.UUID, 
                                peer.UdpAddr, peer.RxOkay, peer.TxOkay, peer.TxError)
        }
        fmt.Fprintf(w, body)
        return
    default:
        http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
        return 
    }
}

func (this *Http) Mac(w http.ResponseWriter, r *http.Request) {
    if (!this.IsAuth(w, r)) {
        return
    }

    switch (r.Method) {
    case "GET":  
        body := "address,remoteaddr,uptime\n"
        for entry := range this.Point.GetMacs() {
            if entry == nil {
                break
            }
            body += fmt.Sprintf("%s, %s, %d\n", entry.Addr, entry.Peer.UdpAddr, entry.Peer.UpTime())
        }
        fmt.Fprintf(w, body)
        return
    default:
        http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
        return 
    }
}