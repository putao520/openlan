package config

import (
	"fmt"
	"github.com/danieldin95/openlan-go/libol"
	"os"
	"strings"
)

type Log struct {
	File    string `json:"file,omitempty" yaml:"file,omitempty"`
	Verbose int    `json:"level,omitempty" yaml:"level,omitempty"`
}

type Http struct {
	Listen string `json:"listen,omitempty" yaml:"listen,omitempty"`
	Public string `json:"public,omitempty" yaml:"public,omitempty"`
}

func RightAddr(listen *string, port int) {
	values := strings.Split(*listen, ":")
	if len(values) == 1 {
		*listen = fmt.Sprintf("%s:%d", values[0], port)
	}
}

func GetAlias() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return libol.GenToken(13)
}
