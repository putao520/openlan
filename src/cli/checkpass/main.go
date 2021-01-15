package main

import (
	"bufio"
	"github.com/danieldin95/openlan-go/src/cli/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"os"
	"path"
	"strings"
)

func findPass(c *config.Switch, username string) string {
	network := "default"
	shortname := username
	if strings.Contains(username, "@") {
		shortname = strings.SplitN(username, "@", 2)[0]
		network = strings.SplitN(username, "@", 2)[1]
	}
	// Password from network.
	for _, net := range c.Network {
		if net.Name != network {
			continue
		}
		for _, auth := range net.Password {
			if auth.Username != shortname {
				continue
			}
			return auth.Password
		}
	}
	// Password from file
	if reader, err := os.Open(c.Password); err == nil {
		defer reader.Close()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			columns := strings.SplitN(line, ":", 4)
			if len(columns) < 2 {
				continue
			}
			if username != columns[0] {
				continue
			}
			return columns[1]
		}
	}
	return ""
}

func main() {
	proc := os.Args[0]
	user := os.Getenv("username")
	pass := os.Getenv("password")

	c := config.NewSwitch()
	passTrue := findPass(c, user)

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
