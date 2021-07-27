package cmd

import (
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/urfave/cli/v2"
)

type Lease struct {
	Cmd
}

func (u Lease) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/lease"
	} else {
		return prefix + "/api/lease/" + name
	}
}

func (u Lease) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "uuid"}} {{ps -16 "alias"}} {{ ps -16 "address" }} {{ps -22 "client"}} {{ps -8 "network"}} {{ ps -6 "type"}}
{{- range . }}
{{ps -16 .UUID}} {{ps -16 .Alias}} {{ ps -16 .Address}} {{ps -22 .Client}} {{ps -8 .Network}} {{ ps -6 .Type}}
{{- end }}
`
}

func (u Lease) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Lease
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Lease) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "lease",
		Aliases: []string{"le"},
		Usage:   "DHCP address lease",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all devices",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
