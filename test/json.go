package main

import (
	"encoding/json"
	"fmt"
	"net"
)

type Hi struct {
	Name string
}

type HardwareAddr struct {
	net.HardwareAddr
}

func (this HardwareAddr) MarshalText() ([]byte, error) {
	if len([]byte(this.HardwareAddr)) == 0 {
		return []byte(""), nil
	}

	return []byte(this.String()), nil
}

func (this *HardwareAddr) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*this = HardwareAddr{nil}
		return nil
	}

	s := string(text)
	x, err := net.ParseMAC(s)
	if err != nil {
		return &net.ParseError{Type: "Hardware address", Text: s}
	}

	*this = HardwareAddr{x}
	return nil
}

type Test struct {
	Username string       `json:"Password,omitempty"`
	Password string       `json:"Password,omit"`
	HwAddr   HardwareAddr `json:"HwAddr"`
	Hi       int          `json:"Hi,string"`
}

func main() {
	t := Test{
		Username: "hi",
		Password: "daniel",
		Hi:       0x21,
	}

	hw, _ := net.ParseMAC("2a:60:84:bd:fe:50")
	t.HwAddr = HardwareAddr{hw}

	str, err := json.Marshal(t)
	fmt.Println(string(str), err)

	o := &Test{}

	err = json.Unmarshal([]byte(str), o)
	fmt.Println(o, err)
}
