package cmd

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

type User struct {
	Cmd
}

func (u User) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/user"
	} else {
		return prefix + "/api/user/" + name
	}
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

var tmpl = `# total {{ len . }}
{{ps -13 "username"}} {{ps -13 "network"}} {{ps -16 "password"}} {{ps -6 "role"}}
{{- range . }}
{{ps -13 .Name}} {{ps -13 .Network}} {{ps -16 .Password}} {{ps -6 .Role}}
{{- end }}
`

func (u User) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
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
	return u.Output(users, c.String("format"), tmpl)
}

func (u User) Get(c *cli.Context) error {
	fullName := c.String("name")
	if n := c.String("network"); n != "" {
		fullName += "@" + n
	}
	url := u.Url(c.String("url"), fullName)
	client := Client{
		Auth: libol.Auth{
			Username: c.String("token"),
		},
	}
	request := client.NewRequest(url)
	users := []schema.User{{}}
	if err := client.GetJSON(request, &users[0]); err != nil {
		return err
	}
	return u.Output(users, c.String("format"), tmpl)
}

func (u User) Check(c *cli.Context) error {
	netFromO := c.String("network")
	nameFromE := c.String("name")
	passFromE := c.String("password")
	if nameFromE == "" {
		nameFromE = os.Getenv("username")
		passFromE = os.Getenv("password")
	}
	netFromE := "default"
	if strings.Contains(nameFromE, "@") {
		netFromE = strings.Split(nameFromE, "@")[1]
	}
	fullName := nameFromE
	if !strings.Contains(nameFromE, "@") {
		fullName = nameFromE + "@" + netFromE
	}
	if netFromO != "" && netFromE != netFromO {
		return libol.NewErr("wrong: zo=%s, us=%s", netFromO, nameFromE)
	}
	url := u.Url(c.String("url"), fullName)
	client := Client{
		Auth: libol.Auth{
			Username: c.String("token"),
		},
	}
	request := client.NewRequest(url)
	var user schema.User
	if err := client.GetJSON(request, &user); err != nil {
		return err
	}
	if user.Password == passFromE {
		fmt.Printf("success: us=%s\n", nameFromE)
		return nil
	}
	return libol.NewErr("wrong: us=%s, pa=%s", nameFromE, passFromE)
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
					&cli.StringFlag{Name: "network"},
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
					&cli.StringFlag{Name: "network"},
					&cli.StringFlag{Name: "name"},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all user",
				Aliases: []string{"ls"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "network"},
				},
				Action: u.List,
			},
			{
				Name:  "get",
				Usage: "Get an user",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
					&cli.StringFlag{Name: "network"},
				},
				Action: u.Get,
			},
			{
				Name:    "check",
				Usage:   "Check an user",
				Aliases: []string{"co"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
					&cli.StringFlag{Name: "password"},
					&cli.StringFlag{Name: "network"},
				},
				Action: u.Check,
			},
		},
	})
}
