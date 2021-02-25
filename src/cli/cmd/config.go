package cmd

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Cmd
}

func (u Config) Url(prefix, name string) string {
	return prefix + "/api/config"
}

func (u Config) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	client := Client{
		Auth: libol.Auth{
			Username: c.String("token"),
		},
	}
	request := client.NewRequest(url)
	if data, err := client.GetBody(request); err == nil {
		fmt.Println(string(data))
		return nil
	} else {
		return err
	}
}

func (u Config) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "config",
		Aliases: []string{"cfg"},
		Usage:   "Configuration",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all configuration",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
