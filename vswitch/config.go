package vswitch

import (
    "flag"
    "strings"
    "fmt"
    "bytes"
    "os"
    "io/ioutil"
    "encoding/json"

    "github.com/lightstar-dev/openlan-go/libol"
    "github.com/lightstar-dev/openlan-go/point"
)

type Config struct {
    TcpListen string `json:"listen"`
    Verbose int `json:"verbose"`
    HttpListen string `json:"http"`
    Ifmtu int `json:"ifMtu"`
    Ifaddr string `json:"ifAddr"`
    Brname string `json:"ifBridge"`
    Token string `json:"adminToken"`
    TokenFile string `json:"adminFile"`
    Password string `json:"authFile"`
    Redis RedisConfig `json:"redis"`

    Links []*point.Config `json:"links"`
    saveFile string
}

type RedisConfig struct {
    Enable bool `json:"enable"`
    Addr string `json:"addr"`
    Auth string `json:"auth"`
    Db int `json:"database"`
}

var Default = Config {
    Brname: "",
    Verbose: 0,
    HttpListen: "0.0.0.0:10000",
    TcpListen: "0.0.0.0:10002",
    Token: "",
    TokenFile: ".vswitch.token",
    Password: ".password",
    Ifmtu: 1518,
    Ifaddr: "",
    Redis: RedisConfig {
        Addr: "127.0.0.1",
        Auth: "",
        Db: 0,
        Enable: false,
    },
    saveFile: ".vswitch.json", 
    Links: nil,
}

func RightAddr(listen *string, port int) {
    values := strings.Split(*listen, ":")
    if len(values) == 1 {
        *listen = fmt.Sprintf("%s:%d", values[0], port)
    }
}

func NewConfig() (this *Config) {
    this = &Config {
        Redis: Default.Redis,
    }

    flag.IntVar(&this.Verbose, "verbose", Default.Verbose, "open verbose")
    flag.StringVar(&this.HttpListen, "http:addr", Default.HttpListen,  "the http listen on")
    flag.StringVar(&this.TcpListen, "vs:addr", Default.TcpListen,  "the server listen on")
    flag.StringVar(&this.Token, "admin:token", Default.Token, "Administrator token")
    flag.StringVar(&this.TokenFile, "admin:file", Default.TokenFile, "The file administrator token saved to")
    flag.StringVar(&this.Password, "auth:file", Default.Password, "The file password loading from.")
    flag.IntVar(&this.Ifmtu, "if:mtu", Default.Ifmtu, "the interface MTU include ethernet")
    flag.StringVar(&this.Ifaddr, "if:addr", Default.Ifaddr, "the interface address")
    flag.StringVar(&this.Brname, "if:br", Default.Brname,  "the bridge name")
    flag.StringVar(&this.saveFile, "conf", Default.SaveFile(), "The configuration file")

    flag.Parse()

    this.Right()
    this.Load()
    this.Save(fmt.Sprintf("%s.cur", this.saveFile))

    str, err := this.Marshal(false)
    if err != nil { 
        libol.Error("NewConfig.json error: %s" , err) 
    }
    libol.Info("NewConfig.json: %s", str)

    return
}

func (this *Config) Right() {
    RightAddr(&this.TcpListen, 10002)
    RightAddr(&this.HttpListen, 10082)

    //TODO reset zero value to default 
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

func (this *Config) SaveFile() string {
    return this.saveFile
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

    contents, err := ioutil.ReadFile(this.saveFile); 
    if err != nil {
        libol.Error("Config.Load: file:%s %s", this.saveFile, err)
        return err
        
    }
    
    if err := json.Unmarshal([]byte(contents), this); err != nil {
        libol.Error("Config.Load: %s", err)
        return err
    }

    if this.Links != nil {
        for _, link := range this.Links {
            link.Right()
        }
    }

    //libol.Debug("Config.Load %s", this)
    return nil
}
