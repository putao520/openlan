package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

type User struct {
}

func (u User) Add(c *cli.Context) error {
	fmt.Println("flags:", c.String("network"),
		c.String("name"), c.String("password"),
		c.String("role"))
	return nil
}

func (u User) Remove(c *cli.Context) error {
	fmt.Println("removed: ", c.Args().First())
	return nil
}

func (u User) List(c *cli.Context) error {
	fmt.Println("list: ", c.Args().First())
	fmt.Println(c.String("url"), c.String("token"))
	return nil
}

func (u User) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "user",
		Aliases: []string{"u"},
		Usage:   "User authentication",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a new user",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network", Value: "default"},
					&cli.StringFlag{Name: "name"},
					&cli.StringFlag{Name: "password"},
					&cli.StringFlag{Name: "role", Value: "guest"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove an existing user",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network", Value: "default"},
					&cli.StringFlag{Name: "name"},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all user",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network", Value: ""},
				},
				Action: u.List,
			},
		},
	})
}
