package cmd

import (
	"fmt"
	"github.com/danieldin95/openlan/src/libol"
	"github.com/danieldin95/openlan/src/schema"
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
	username := c.String("name")
	user := &schema.User{
		Name:     username,
		Password: c.String("password"),
		Role:     c.String("role"),
	}
	if user.Name == "" {
		return libol.NewErr("name is empty")
	}
	if user.Password == "" {
		return libol.NewErr("password is empty")
	}
	if !strings.Contains(username, "@") {
		return libol.NewErr("name not contains network")
	}
	values := strings.SplitN(username, "@", 2)
	user.Name = values[0]
	user.Network = values[1]
	url := u.Url(c.String("url"), user.Name)
	clt := u.NewHttp(c.String("token"))
	if err := clt.PostJSON(url, user); err != nil {
		return err
	}
	return nil
}

func (u User) Remove(c *cli.Context) error {
	username := c.String("name")
	url := u.Url(c.String("url"), username)
	clt := u.NewHttp(c.String("token"))
	if err := clt.DeleteJSON(url, nil); err != nil {
		return err
	}
	return nil
}

func (u User) Tmpl() string {
	return `# total {{ len . }}
{{ps -24 "username"}} {{ps -24 "password"}} {{ps -6 "role"}}
{{- range . }}
{{p2 -24 "%s@%s" .Name .Network}} {{ps -24 .Password}} {{ps -6 .Role}}
{{- end }}
`
}

func (u User) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.User
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u User) Get(c *cli.Context) error {
	username := c.String("name")
	url := u.Url(c.String("url"), username)
	client := u.NewHttp(c.String("token"))
	items := []schema.User{{}}
	if err := client.GetJSON(url, &items[0]); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
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
	alias := ""
	if ip, ok := os.LookupEnv("untrusted_ip"); ok {
		alias = ip + ":" + os.Getenv("untrusted_port")
	}
	url := u.Url(c.String("url"), fullName)
	url += "/check"
	client := u.NewHttp(c.String("token"))
	user := &schema.User{
		Name:     fullName,
		Password: passFromE,
		Alias:    alias,
	}
	if err := client.PostJSON(url, user); err != nil {
		return err
	} else {
		fmt.Printf("success: us=%s\n", nameFromE)
		return nil
	}
}

func (u User) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "user",
		Aliases: []string{"us"},
		Usage:   "User authentication",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a new user",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
					&cli.StringFlag{Name: "password", Value: libol.GenRandom(24)},
					&cli.StringFlag{Name: "role", Value: "guest"},
				},
				Action: u.Add,
			},
			{
				Name:    "remove",
				Usage:   "Remove an existing user",
				Aliases: []string{"rm"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name"},
				},
				Action: u.Remove,
			},
			{
				Name:    "list",
				Usage:   "Display all users",
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
