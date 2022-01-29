package main

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/danieldin95/openlan/cmd/conf"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	confSockFile   = "unix:/var/openlan/confd.sock"
	confDatabase   = "OpenLAN_Switch"
	adminTokenFile = "/etc/openlan/switch/token"
)

type App struct {
	Token    string
	Url      string
	Conf     string
	Database string
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
		&cli.StringFlag{
			Name:    "conf",
			Aliases: []string{"c"},
			Usage:   "confd server connection",
			Value:   a.Conf,
		},
		&cli.StringFlag{
			Name:    "database",
			Aliases: []string{"d"},
			Usage:   "confd database",
			Value:   a.Database,
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

	server := os.Getenv("OL_CONF")
	if server == "" {
		server = confSockFile
	}
	database := os.Getenv("OL_DATABASE")
	if database == "" {
		database = confDatabase
	}
	app := App{
		Url:      url,
		Token:    token,
		Conf:     server,
		Database: database,
	}.New()

	app.Commands = api.User{}.Commands(app)
	app.Commands = api.ACL{}.Commands(app)
	app.Commands = api.Device{}.Commands(app)
	app.Commands = api.Lease{}.Commands(app)
	app.Commands = api.Config{}.Commands(app)
	app.Commands = api.Point{}.Commands(app)
	app.Commands = api.VPNClient{}.Commands(app)
	app.Commands = api.Link{}.Commands(app)
	app.Commands = api.Server{}.Commands(app)
	app.Commands = api.Network{}.Commands(app)
	app.Commands = api.PProf{}.Commands(app)
	app.Commands = api.Esp{}.Commands(app)
	app.Commands = api.VxLAN{}.Commands(app)
	app.Commands = api.State{}.Commands(app)
	app.Commands = api.Policy{}.Commands(app)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	if err := conf.Open(server, database); err == nil {
		conf.List()
	}
}
