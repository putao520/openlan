package cmd

import (
	"github.com/danieldin95/openlan/src/schema"
	"github.com/urfave/cli/v2"
)

type Link struct {
	Cmd
}

func (u Link) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/link"
	} else {
		return prefix + "/api/link/" + name
	}
}

func (u Link) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "uuid"}} {{ps -8 "alive"}} {{ ps -8 "device" }} {{ps -8 "user"}} {{ps -22 "server"}} {{ps -8 "network"}} {{ ps -6 "state"}}
{{- range . }}
{{ps -16 .UUID}} {{pt .AliveTime | ps -8}} {{ ps -8 .Device}} {{ps -8 .User}} {{ps -22 .Server}} {{ps -8 .Network}}  {{ ps -6 .State}}
{{- end }}
`
}

func (u Link) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.Link
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Link) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "link",
		Aliases: []string{"ln"},
		Usage:   "Link connect to others",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all links",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
