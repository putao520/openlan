package api

import (
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/urfave/cli/v2"
	"io/ioutil"
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
	cli *cli.App
}

func (a *App) Flags() []cli.Flag {
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

func (a *App) New() *cli.App {
	app := &cli.App{
		Usage:    "OpenLAN switch utility",
		Flags:    a.Flags(),
		Commands: []*cli.Command{},
		Before:   a.Before,
		After:    a.After,
	}
	a.cli = app
	return a.cli
}

func (a *App) Before(c *cli.Context) error {
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

func (a *App) After(c *cli.Context) error {
	return nil
}

func (a *App) Command(cmd *cli.Command) {
	a.cli.Commands = append(a.cli.Commands, cmd)
}

func (a *App) Run(args []string) error {
	return a.cli.Run(args)
}
