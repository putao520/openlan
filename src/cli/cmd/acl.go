package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

type ACL struct {
}

func (u ACL) Add(c *cli.Context) error {
	fmt.Println("flags:", c.String("network"), c.String("name"),
		c.String("password"), c.String("role"))
	return nil
}

func (u ACL) Remove(c *cli.Context) error {
	fmt.Println("removed task template: ", c.Args().First())
	return nil
}

func (u ACL) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "acl",
		Aliases: []string{"a"},
		Usage:   "Access control list",
		Subcommands: []*cli.Command{
			{
				Name:   "add",
				Usage:  "Add a new template",
				Action: u.Add,
			},
			{
				Name:   "remove",
				Usage:  "Remove an existing template",
				Action: u.Remove,
			},
		},
	})
}
