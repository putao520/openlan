package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

type User struct {
}

func (u User) Add(c *cli.Context) error {
	fmt.Println("flags:", c.String("name"),
		c.String("password"), c.String("role"))
	return nil
}

func (u User) Remove(c *cli.Context) error {
	fmt.Println("removed task template: ", c.Args().First())
	return nil
}

func (u User) Command(app *cli.App) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    "user",
		Aliases: []string{"u"},
		Usage:   "options for user",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add a new user",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}},
					&cli.StringFlag{Name: "password", Aliases: []string{"p"}},
					&cli.StringFlag{Name: "role", Aliases: []string{"r"}, Value: "guest"},
				},
				Action: u.Add,
			},
			{
				Name:   "remove",
				Usage:  "remove an existing template",
				Action: u.Remove,
			},
		},
	})
}
