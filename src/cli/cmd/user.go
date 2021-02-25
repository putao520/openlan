package cmd

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/urfave/cli/v2"
)

type User struct {
}

func (u User) Url(prefix string) string {
	return prefix + "/api/user"
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
	url := u.Url(c.String("url"))
	client := Client{
		Auth: libol.Auth{
			Username: c.String("token"),
		},
	}
	request := client.NewRequest(url)
	var users []schema.User
	if err := client.GetJSON(request, &users); err != nil {
		return err
	}
	if out, err := libol.Marshal(users, true); err == nil {
		fmt.Println(string(out))
	}
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
