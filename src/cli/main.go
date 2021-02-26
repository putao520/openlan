package main

import (
	"github.com/danieldin95/openlan-go/src/cli/cmd"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	adminTokenFile = "/etc/openlan/switch/token"
)

type App struct {
	Token string
	Url   string
}

func (a App) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "token",
			Aliases: []string{"t"},
			Usage:   "admin token",
			Value:   a.Token,
		},
		&cli.StringFlag{
			Name:    "url",
			Aliases: []string{"l"},
			Usage:   "server url",
			Value:   a.Url,
		},
		&cli.StringFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Usage:   "output format: json, yaml",
			Value:   "table",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Enable verbose",
			Value:   false,
		},
	}
}

func (a App) New() *cli.App {
	return &cli.App{
		Usage:    "OpenLAN switch utility",
		Flags:    a.Flags(),
		Commands: []*cli.Command{},
		Before:   a.Before,
		After:    a.After,
	}
}

func (a App) Before(c *cli.Context) error {
	if c.Bool("verbose") {
		libol.SetLogger("", libol.DEBUG)
	}
	token := c.String("token")
	if token == "" {
		if data, err := ioutil.ReadFile(adminTokenFile); err == nil {
			token = strings.TrimSpace(string(data))
		}
		_ = c.Set("token", token)
	}
	return nil
}

func (a App) After(c *cli.Context) error {
	return nil
}

func main() {
	url := os.Getenv("OL_URL")
	if url == "" {
		url = "https://localhost:10000"
	}
	token := os.Getenv("OL_TOKEN")

	app := App{Url: url, Token: token}.New()
	app.Commands = cmd.User{}.Commands(app)
	app.Commands = cmd.ACL{}.Commands(app)
	app.Commands = cmd.Device{}.Commands(app)
	app.Commands = cmd.Lease{}.Commands(app)
	app.Commands = cmd.Config{}.Commands(app)
	app.Commands = cmd.Point{}.Commands(app)
	app.Commands = cmd.OvClient{}.Commands(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
