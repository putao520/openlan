package cmd

import (
	"github.com/urfave/cli/v2"
)

type ACL struct {
	Cmd
}

func (u ACL) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/acl"
	} else {
		return prefix + "/api/acl/" + name
	}
}

func (u ACL) Add(c *cli.Context) error {
	return nil
}

func (u ACL) Remove(c *cli.Context) error {
	return nil
}

func (u ACL) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:  "acl",
		Usage: "Access control list",
		Subcommands: []*cli.Command{
			{
				Name:   "add",
				Usage:  "Add a new acl",
				Action: u.Add,
			},
			{
				Name:   "remove",
				Usage:  "Remove an existing acl",
				Action: u.Remove,
			},
		},
	})
}
