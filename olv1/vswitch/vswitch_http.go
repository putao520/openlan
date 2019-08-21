package vswitch

import (
    "os"
    "fmt"
    "html"
    "log"
    "time"
    "math/rand"
    "io/ioutil"
    "net/http"
    "encoding/json"
)

type OpeHttp struct {
    wroker *OpeWroker
    listen string
    adminToken string
    adminFile string
}

func getToken(n int) string {
    letters := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
    buf := make([]byte, n)

    rand.Seed(time.Now().UnixNano())

    for i := range buf {
        buf[i] = letters[rand.Int63() % int64(len(letters))]
    }

    return string(buf)
}

func NewOpeHttp(wroker *OpeWroker, c *Config)(this *OpeHttp) {
    token := c.Token
    if token == "" {
        token = getToken(16)
    }
    this = &OpeHttp {
        wroker: wroker,
        listen: c.HttpListen,
        adminToken: token,
        adminFile: c.TokenFile,
    }

    this.SaveToken()
    http.HandleFunc("/", this.Index)
    http.HandleFunc("/hi", this.Hi)
    http.HandleFunc("/user", this.User)

    return 
}

func (this *OpeHttp) SaveToken() error {
    log.Printf("OpeHttp.SaveToken: AdminToken: %s", this.adminToken)

    f, err := os.OpenFile(this.adminFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
    defer f.Close()
    if err != nil {
        log.Printf("Error| OpeHttp.SaveToken: %s", err)
        return err
    }

    if _, err := f.Write([]byte(this.adminToken)); err != nil {
        log.Printf("Error| OpeHttp.SaveToken: %s", err)
        return err
    }

    return nil
}

func (this *OpeHttp) GoStart() error {
    log.Printf("NewHttp on %s", this.listen)
    if err := http.ListenAndServe(this.listen, nil); err != nil {
        log.Printf("Error| OpeHttp.GoStart on %s: %s", this.listen, err)
        return err
    }
    return nil
}

func (this *OpeHttp) IsAuth(w http.ResponseWriter, r *http.Request) bool {
    token, pass, ok := r.BasicAuth()
    if this.wroker.IsVerbose() {
        log.Printf("token: %s, pass: %s", token, pass)
    }
    if !ok  || token != this.adminToken {
        w.Header().Set("WWW-Authenticate", "Basic")
        http.Error(w, "Authorization Required.", 401)
        return false
    }

    return true
}

func (this *OpeHttp) Hi(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi %s %q", r.Method, html.EscapeString(r.URL.Path))

    for name, headers := range r.Header {
        for _, h := range headers {
            log.Printf("Info| OpeHttp.Hi %v: %v", name, h)
        }
    }
}

func (this *OpeHttp) Index(w http.ResponseWriter, r *http.Request) {
    if (!this.IsAuth(w, r)) {
        return
    }

    switch (r.Method) {
    case "GET":  
        body := "uptime, remoteaddr, device, receipt, transmission, error\n"
        for client, ifce := range this.wroker.Clients {
            body += fmt.Sprintf("%d, %s, %s, %d, %d, %d\n", 
                                client.UpTime(), client.GetAddr(), ifce.Name(),
                                client.RxOkay, client.TxOkay, client.TxError)
        }
        fmt.Fprintf(w, body)
    default:
        http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
        return 
    }
}

func (this *OpeHttp) User(w http.ResponseWriter, r *http.Request) {
    if (!this.IsAuth(w, r)) {
        return
    }

    switch (r.Method) {
    case "GET":
        pagesJson, err := json.Marshal(this.wroker.Users)
        if err != nil {
            fmt.Fprintf(w, fmt.Sprintf("Error| OpeHttp.User: %s", err))
            return
        }

        fmt.Fprintf(w, string(pagesJson))
    case "POST":
        defer r.Body.Close()
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            http.Error(w, fmt.Sprintf("Error| OpeHttp.User: %s", err), 400)
            return
        }

        user := &User {}
        if err := json.Unmarshal([]byte(body), user); err != nil {
            http.Error(w, fmt.Sprintf("Error| OpeHttp.User: %s", err), 400)
            return
        }

        if user.Name != "" {
            this.wroker.Users[user.Name] = user
        } else if (user.Token != "") {
            this.wroker.Users[user.Token] = user
        }

        fmt.Fprintf(w, "Saved it.")
    default:
        http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
    }
}