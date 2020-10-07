package main

import (
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"os"
	"path"
	"strings"
)

func main() {
	proc := os.Args[0]
	user := os.Getenv("username")
	pass := os.Getenv("password")

	name := user
	network := "default"
	if strings.Contains(user, "@") {
		name = strings.SplitN(user, "@", 2)[0]
		network = strings.SplitN(user, "@", 2)[1]
	}
	passTrue := ""
	c := config.NewSwitch()
	for _, net := range c.Network {
		if net.Name != network {
			continue
		}
		for _, auth := range net.Password {
			if auth.Username != name {
				continue
			}
			passTrue = auth.Password
			break
		}
		break
	}
	logDir := path.Dir(c.Log.File)
	logFile := logDir + "/" + path.Base(proc) + ".log"
	libol.SetLogger(logFile, c.Log.Verbose)
	if passTrue == "" {
		libol.Warn("notExisted: username=%s", user)
		os.Exit(1)
	} else if pass == passTrue {
		libol.Info("success: username=%s, and password=%s", user, pass)
		os.Exit(0)
	} else {
		libol.Warn("notRight: username=%s, and password=%s", user, pass)
		os.Exit(1)
	}
}
