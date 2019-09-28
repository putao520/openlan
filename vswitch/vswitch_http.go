package vswitch

import (
    "os"
    "fmt"
    "html"
    "time"
    "math/rand"
    "io/ioutil"
    "net/http"
    "encoding/json"

    "github.com/lightstar-dev/openlan-go/libol"
    "github.com/lightstar-dev/openlan-go/point"
)

type VSwitchHttp struct {
    wroker *VSwitchWroker
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

func NewVSwitchHttp(wroker *VSwitchWroker, c *Config)(this *VSwitchHttp) {
    this = &VSwitchHttp {
        wroker: wroker,
        listen: c.HttpListen,
        adminToken: c.Token,
        adminFile: c.TokenFile,
    }

    if this.adminToken == "" {
        this.LoadToken()
    }

    if this.adminToken == "" {
        this.adminToken = getToken(13)
    }

    this.SaveToken()
    http.HandleFunc("/", this.Index)
    http.HandleFunc("/hello", this.Hello)
    http.HandleFunc("/api/user", this._User)
    http.HandleFunc("/api/neighbor", this._Neighbor)
    http.HandleFunc("/api/link", this._Link)

    return 
}

func (this *VSwitchHttp) SaveToken() error {
    libol.Info("VSwitchHttp.SaveToken: AdminToken: %s", this.adminToken)

    f, err := os.OpenFile(this.adminFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
    defer f.Close()
    if err != nil {
        libol.Error("VSwitchHttp.SaveToken: %s", err)
        return err
    }

    if _, err := f.Write([]byte(this.adminToken)); err != nil {
        libol.Error("VSwitchHttp.SaveToken: %s", err)
        return err
    }

    return nil
}

func (this *VSwitchHttp) LoadToken() error {
    if _, err := os.Stat(this.adminFile); os.IsNotExist(err) {
        libol.Info("VSwitchHttp.LoadToken: file:%s does not exist", this.adminFile)
        return nil
    }

    contents, err := ioutil.ReadFile(this.adminFile); 
    if err != nil {
        libol.Error("VSwitchHttp.LoadToken: file:%s %s", this.adminFile, err)
        return err
        
    }
    
    this.adminToken = string(contents)
    return nil
}

func (this *VSwitchHttp) GoStart() error {
    libol.Debug("NewHttp on %s", this.listen)

    //hfs := http.FileServer(http.Dir("."))
    if err := http.ListenAndServe(this.listen, nil); err != nil {
        libol.Error("VSwitchHttp.GoStart on %s: %s", this.listen, err)
        return err
    }
    return nil
}

func (this *VSwitchHttp) IsAuth(w http.ResponseWriter, r *http.Request) bool {
    token, pass, ok := r.BasicAuth()
    if this.wroker.IsVerbose() {
        libol.Debug("VSwitchHttp.IsAuth token: %s, pass: %s", token, pass)
    }
    if !ok  || token != this.adminToken {
        w.Header().Set("WWW-Authenticate", "Basic")
        http.Error(w, "Authorization Required.", 401)
        return false
    }

    return true
}

func (this *VSwitchHttp) Hello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello %s %q", r.Method, html.EscapeString(r.URL.Path))

    for name, headers := range r.Header {
        for _, h := range headers {
            libol.Info("VSwitchHttp.Hello %v: %v", name, h)
        }
    }
}

func (this *VSwitchHttp) Index(w http.ResponseWriter, r *http.Request) {
    if (!this.IsAuth(w, r)) {
        return
    }

    switch (r.Method) {
    case "GET":  
        body := fmt.Sprintf("# uptime: %d\n", this.wroker.UpTime())
        body += "\n"
        body += "# point accessed to this vswith.\n"
        body += "uptime, remote, device, receipt, transmis, error\n"
        for p := range this.wroker.ListPoint() {
            if p == nil {
                break
            }

            client, ifce := p.Client, p.Device
            body += fmt.Sprintf("%d, %s, %s, %d, %d, %d\n", 
                                client.UpTime(), client.Addr, ifce.Name(),
                                client.RxOkay, client.TxOkay, client.TxError)
        }

        body += "\n"
        body += "# neighbor we discovered on this vswitch.\n"
        body += "uptime, ethernet, address, remote\n"
        for n := range this.wroker.Neighbor.ListNeighbor() {
            if n == nil {
                break
            }
            
            body += fmt.Sprintf("%d, %s, %s, %s\n", 
                                n.UpTime(), n.HwAddr, n.IpAddr, n.Client)
        }

        body += "\n"
        body += "# link which connect to other vswitch.\n"
        body += "uptime, bridge, device, remote\n"
        for p := range this.wroker.ListLink() {
            if p == nil {
                break
            }
            
            body += fmt.Sprintf("%d, %s, %s, %s\n", 
                                p.Client.UpTime(), p.Brname, p.Ifname, p.Client.Addr)
        }
        
        fmt.Fprintf(w, body)
    default:
        http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
        return 
    }
}

func (this *VSwitchHttp) Marshal(v interface {}) (string, error) {
    str , err := json.Marshal(v)
    if err != nil { 
        libol.Error("VSwitchHttp.Marsha1: %s" , err)
        return "", err
    }

    return string(str), nil
}

func (this *VSwitchHttp) _User(w http.ResponseWriter, r *http.Request) {
    if (!this.IsAuth(w, r)) {
        return
    }

    switch (r.Method) {
    case "GET":
        users := make([]*User, 0, 1024)
        for u := range this.wroker.ListUser() {
            if u == nil {
                break
            }
            users = append(users, u)
        }

        body, _:= this.Marshal(users)
        fmt.Fprintf(w, body)
    case "POST":
        defer r.Body.Close()
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            http.Error(w, fmt.Sprintf("Error| VSwitchHttp._User: %s", err), 400)
            return
        }

        user := &User {}
        if err := json.Unmarshal([]byte(body), user); err != nil {
            http.Error(w, fmt.Sprintf("Error| VSwitchHttp._User: %s", err), 400)
            return
        }

        this.wroker.AddUser(user)

        fmt.Fprintf(w, "success")
    default:
        http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
    }
}

func (this *VSwitchHttp) _Neighbor(w http.ResponseWriter, r *http.Request) {
    if (!this.IsAuth(w, r)) {
        return
    }

    switch (r.Method) {
    case "GET":  
        neighbors := make([]*Neighbor, 0, 1024)
        for n := range this.wroker.Neighbor.ListNeighbor() {
            if n == nil {
                break
            }
            
            neighbors = append(neighbors, n)
        }
        
        body, _ := this.Marshal(neighbors)
        fmt.Fprintf(w, body)
    default:
        http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
        return 
    }
}

func (this *VSwitchHttp) _Link(w http.ResponseWriter, r *http.Request) {
    if (!this.IsAuth(w, r)) {
        return
    }

    switch (r.Method) {
    case "GET":
        links := make([]*point.Point, 0, 1024)
        for l := range this.wroker.ListLink() {
            if l == nil {
                break
            }
            links = append(links, l)
        }
        body, _ := this.Marshal(links)
        fmt.Fprintf(w, body)
    case "POST":
        defer r.Body.Close()
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            http.Error(w, fmt.Sprintf("Error| VSwitchHttp._Link: %s", err), 400)
            return
        }

        c := &point.Config {}
        if err := json.Unmarshal([]byte(body), c); err != nil {
            http.Error(w, fmt.Sprintf("Error| VSwitchHttp._Link: %s", err), 400)
            return
        }
        
        c.Default()
        this.wroker.AddLink(c)

        fmt.Fprintf(w, "success")
    default:
        http.Error(w, fmt.Sprintf("Not support %s", r.Method), 400)
    }
}


