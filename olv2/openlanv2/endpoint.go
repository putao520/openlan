package openlanv2

import (
    "net"
    "encoding/json"
    "strings"
    "time"
    "fmt"
)

type Endpoint struct {
    Name string `json:"name"`
    UUID string `json:"uuid"`
    Password string `json:"password"`
    Network string
    UdpAddr *net.UDPAddr
    TxOkay uint64
    RxOkay uint64
    TxError uint64
    Expired int
    //
    updateTime int64
    createTime int64
}

func GetNetwork(name string) (string) {
    values := strings.Split(name, "@")
    if len(values) == 2 {
        return values[1]
    }
    return ""
}

func NewEndpoint(name string, uuid string) (this *Endpoint) {
    this = &Endpoint {
        Name: name, 
        UUID: uuid,
        createTime: time.Now().Unix(),
        updateTime: time.Now().Unix(),
        Expired: 60, //Default expired time is 60s. 
    }

    this.Network = GetNetwork(this.Name)
    return
}

func NewEndpointFromJson(data string) (this *Endpoint, err error) {
    this = &Endpoint {
        createTime: time.Now().Unix(),
        updateTime: time.Now().Unix(),
        Expired: 60, //Default expired time is 60s. 
    }
    if err = json.Unmarshal([]byte(data), this); err != nil {
        return
    }

    this.Network = GetNetwork(this.Name)
    return
}

func (this *Endpoint) Equal(obj *Endpoint) bool {
    return this.Name == obj.Name && this.UUID == obj.UUID 
}

func (this *Endpoint) ToJson() (string, error) {
    data, err := json.Marshal(this)
    return string(data), err
}

func (this *Endpoint) Update(from *Endpoint) bool {
    this.UUID = from.UUID
    this.UdpAddr = from.UdpAddr
    this.updateTime = time.Now().Unix()
    this.RxOkay++

    return false
}

func (this *Endpoint) UpTime() int64 {
    return time.Now().Unix() - this.createTime
}

func (this *Endpoint) String() string {
    return fmt.Sprintf("%s,%s,%s", this.Name, this.UUID, this.UdpAddr)
}

func (this *Endpoint) UpdateTime() int64 {
    return time.Now().Unix() - this.updateTime
}

func (this *Endpoint) IsExpired() bool {
    return this.UpdateTime() > int64(this.Expired)
}