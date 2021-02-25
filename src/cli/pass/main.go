package main

import (
	"bufio"
	"flag"
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"os"
	"path"
	"strings"
)

func splitUsername(username string) (string, string) {
	network := "default"
	shortname := username
	if strings.Contains(username, "@") {
		shortname = strings.SplitN(username, "@", 2)[0]
		network = strings.SplitN(username, "@", 2)[1]
	}
	return shortname, network
}

func findPass(c *PassConfig, username string) string {
	network, shortname := splitUsername(username)
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

type PassConfig struct {
	*config.Switch
	Zone string
}

func NewPassConfig() *PassConfig {
	pa := &PassConfig{
		Switch: &config.Switch{},
	}
	pa.Flags()
	pa.Parse()
	pa.Initialize()
	pa.SetLog()
	return pa
}

func (pa *PassConfig) SetLog() {
	proc := os.Args[0]
	logDir := path.Dir(pa.Log.File)
	logFile := logDir + "/" + path.Base(proc) + ".log"
	libol.SetLogger(logFile, pa.Log.Verbose)
}

func (pa *PassConfig) Flags() {
	pa.Switch.Flags()
	flag.StringVar(&pa.Zone, "zone", "default", "Configure default network")
}

func main() {
	username := os.Getenv("username")
	password := os.Getenv("password")

	c := NewPassConfig()

	_, network := splitUsername(username)
	if c.Zone != network {
		libol.Warn("wrong: zo=%s, us=%s", c.Zone, username)
		os.Exit(1)
	}
	passTo := findPass(c, username)
	if passTo == "" {
		libol.Warn("notExist: us=%s", username)
		os.Exit(1)
	} else if password == passTo {
		libol.Info("success: us=%s", username)
		os.Exit(0)
	} else {
		libol.Warn("wrong: us=%s, pa=%s", username, password)
		os.Exit(1)
	}
}
