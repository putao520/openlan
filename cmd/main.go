package main

import (
	"github.com/danieldin95/openlan/cmd/apiv5"
	"github.com/danieldin95/openlan/cmd/apiv6"
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

var (
	Version  = "v5"
	Url      = "https://localhost:10000"
	Token    = ""
	Server   = confSockFile
	Database = confDatabase
)

type App struct {
}

func (a App) Flags() []cli.Flag {
	var flags []cli.Flag

	if Version == "v5" {
		flags = append(flags,
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "admin token",
				Value:   Token,
			})
		flags = append(flags,
			&cli.StringFlag{
				Name:    "url",
				Aliases: []string{"l"},
				Usage:   "server url",
				Value:   Url,
			})
	}
	flags = append(flags,
		&cli.StringFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Usage:   "output format: json, yaml",
			Value:   "table",
		})
	flags = append(flags,
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "enable verbose",
			Value:   false,
		})
	if Version == "v6" {
		flags = append(flags,
			&cli.StringFlag{
				Name:    "conf",
				Aliases: []string{"c"},
				Usage:   "confd server connection",
				Value:   Server,
			})
		flags = append(flags,
			&cli.StringFlag{
				Name:    "database",
				Aliases: []string{"d"},
				Usage:   "confd database",
				Value:   Database,
			})
	}
	return flags
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

func GetEnv(key, value string) string {
	val := os.Getenv(key)
	if val == "" {
		return value
	}
	return val
}

func main() {
	Version = GetEnv("OL_VERSION", Version)
	Url = GetEnv("OL_URL", Url)
	Token = GetEnv("OL_TOKEN", Token)
	Server = GetEnv("OL_CONF", Server)
	Database = GetEnv("OL_DATABASE", Database)
	app := App{}.New()

	switch Version {
	case "v6":
		if err := apiv6.NewConf(Server, Database, false); err != nil {
			log.Fatal(err)
		}
		app.Commands = apiv6.Switch{}.Commands(app)
	default:
		app.Commands = apiv5.User{}.Commands(app)
		app.Commands = apiv5.ACL{}.Commands(app)
		app.Commands = apiv5.Device{}.Commands(app)
		app.Commands = apiv5.Lease{}.Commands(app)
		app.Commands = apiv5.Config{}.Commands(app)
		app.Commands = apiv5.Point{}.Commands(app)
		app.Commands = apiv5.VPNClient{}.Commands(app)
		app.Commands = apiv5.Link{}.Commands(app)
		app.Commands = apiv5.Server{}.Commands(app)
		app.Commands = apiv5.Network{}.Commands(app)
		app.Commands = apiv5.PProf{}.Commands(app)
		app.Commands = apiv5.Esp{}.Commands(app)
		app.Commands = apiv5.VxLAN{}.Commands(app)
		app.Commands = apiv5.State{}.Commands(app)
		app.Commands = apiv5.Policy{}.Commands(app)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
