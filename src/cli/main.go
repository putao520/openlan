package main

import (
	"github.com/danieldin95/openlan-go/src/cli/cmd"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	adminTokenFile = "/etc/openlan/switch/token"
)

func main() {
	url := os.Getenv("OL_URL")
	if url == "" {
		url = "https://localhost:10000"
	}
	token := os.Getenv("OL_TOKEN")
	if token == "" {
		if data, err := ioutil.ReadFile(adminTokenFile); err == nil {
			token = strings.TrimSpace(string(data))
		}
	}

	app := &cli.App{
		Usage: "OpenLAN switch utility",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "admin token",
				Value:   token,
			},
			&cli.StringFlag{
				Name:    "url",
				Aliases: []string{"l"},
				Usage:   "server url",
				Value:   url,
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "output format: json, yaml",
				Value:   "table",
			},
		},
		Commands: []*cli.Command{},
	}
	app.Commands = cmd.User{}.Commands(app)
	app.Commands = cmd.ACL{}.Commands(app)
	app.Commands = cmd.Device{}.Commands(app)
	app.Commands = cmd.Lease{}.Commands(app)
	app.Commands = cmd.Config{}.Commands(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
