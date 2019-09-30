package point

import (
    "flag"
    "io/ioutil"
    "strings"
    "fmt"
    "os"
    "bytes"
    "encoding/json"

    "github.com/lightstar-dev/openlan-go/libol"    
)

type Config struct {
    Addr string `json:"vsAddr"`
    Auth string `json:"vsAuth"`
    Verbose int `json:"verbose"`
    Ifmtu int `json:"ifMtu"`
    Ifaddr string `json:"ifAddr"`
    Brname string `json:"ifBridge"`
    Iftun bool `json:"ifTun"`
    Ifethsrc string `json:"ifEthSrc"`
    Ifethdst string `json:"ifEthDst"`

    saveFile string
    name string
    password string
}

var Default = Config {
    Addr: "openlan.net",
    Auth: "hi:hi@123$",
    Verbose: 0,
    Ifmtu: 1518,
    Ifaddr: "",
    Iftun: false,
    Brname: "",
    saveFile: ".point.json",
    name: "",
    password: "",
    Ifethdst: "2e:4b:f0:b7:6d:ba",
    Ifethsrc: "",
}

func RightAddr(listen *string, port int) {
    values := strings.Split(*listen, ":")
    if len(values) == 1 {
        *listen = fmt.Sprintf("%s:%d", values[0], port)
    }
}

func NewConfig() (this *Config) {
    this = &Config {}

    flag.StringVar(&this.Addr, "vs:addr", Default.Addr,  "the server connect to")
    flag.StringVar(&this.Auth, "vs:auth", Default.Auth,  "the auth login to")
    flag.IntVar(&this.Verbose, "verbose", Default.Verbose, "open verbose")
    flag.IntVar(&this.Ifmtu, "if:mtu", Default.Ifmtu, "the interface MTU include ethernet")
    flag.StringVar(&this.Ifaddr, "if:addr", Default.Ifaddr, "the interface address")
    flag.StringVar(&this.Brname, "if:br", Default.Brname,  "the bridge name")
    flag.BoolVar(&this.Iftun, "if:tun", Default.Iftun,  "using tun device as interface, otherwise tap")
    flag.StringVar(&this.Ifethdst, "if:ethdst", Default.Ifethdst,  "ethernet destination for tun device")
    flag.StringVar(&this.Ifethsrc, "if:ethsrc", Default.Ifethsrc,  "ethernet source for tun device")
    flag.StringVar(&this.saveFile, "conf", Default.SaveFile(), "The configuration file")

    flag.Parse()

    this.Load()
    this.Default()
    this.Save(fmt.Sprintf("%s.cur", this.saveFile))
    str, err := this.Marshal(false)
    if err != nil { 
        libol.Error("NewConfig.json error: %s" , err) 
    }
    libol.Info("NewConfig.json: %s", str)
    
    return
}

func (this *Config) Default() {
    if this.Auth != "" {
        values := strings.Split(this.Auth, ":")
        this.name = values[0] 
        if len(values) > 1 {
            this.password = values[1]
        }
    }

    RightAddr(&this.Addr, 10002)

    //reset zero value to default 
    if this.Addr == "" {
        this.Addr = Default.Addr
    }
    if this.Auth == "" {
        this.Auth = Default.Auth
    }
    if this.Ifmtu == 0 {
        this.Ifmtu = Default.Ifmtu
    }
    if this.Ifaddr == "" {
        this.Ifaddr = Default.Ifaddr
    }
}

func (this *Config) Name() string {
    return this.name
}

func (this *Config) Password() string {
    return this.password
}

func (this *Config) SaveFile() string {
    return this.saveFile
}

func (this *Config) Marshal(pretty bool) (string, error) {
    str , err := json.Marshal(this)
    if err != nil { 
        libol.Error("NewConfig.json error: %s" , err)
        return "", err
    }

    if !pretty {
        return string(str), nil
    }

    var out bytes.Buffer
    
    if err := json.Indent(&out, str, "", "  "); err != nil {
        return string(str), nil
    }
    
    return out.String(), nil
}

func (this *Config) Save(file string) error {
    if file == "" {
        file = this.saveFile
    }

    f, err := os.OpenFile(file, os.O_RDWR | os.O_TRUNC | os.O_CREATE, 0600)
    defer f.Close()
    if err != nil {
        libol.Error("Config.Save: %s", err)
        return err
    }

    str, err := this.Marshal(true)
    if err != nil { 
        libol.Error("Config.Save error: %s" , err)
        return err
    }

    if _, err := f.Write([]byte(str)); err != nil {
        libol.Error("Config.Save: %s", err)
        return err
    }

    return nil
}

func (this *Config) Load() error {
    if _, err := os.Stat(this.saveFile); os.IsNotExist(err) {
        libol.Info("Config.Load: file:%s does not exist", this.saveFile)
        return nil
    }

    contents, err := ioutil.ReadFile(this.saveFile)
    if err != nil {
        libol.Error("Config.Load: file:%s %s", this.saveFile, err)
        return err

    }

    if err := json.Unmarshal([]byte(contents), this); err != nil {
        libol.Error("Config.Load: %s", err)
        return err
    }

    return nil
}