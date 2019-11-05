package config

import (
	"fmt"
	"github.com/lightstar-dev/openlan-go/libol"
	"os"
	"strings"
)

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
