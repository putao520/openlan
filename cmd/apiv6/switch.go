package apiv6

import (
	"github.com/urfave/cli/v2"
)

type Switch struct {
}

func (u Switch) List(c *cli.Context) error {
	ovs.List()
	return nil
}

func (u Switch) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "switch",
		Aliases: []string{"sw"},
		Usage:   "Global switch",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display switch configuration",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
