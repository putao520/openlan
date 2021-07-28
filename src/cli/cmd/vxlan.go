package cmd

import (
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/urfave/cli/v2"
)

type VxLAN struct {
	Cmd
}

func (u VxLAN) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/vxlan"
	} else {
		return prefix + "/api/vxlan/" + name
	}
}

func (u VxLAN) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "uuid"}} {{ps -8 "alive"}} {{ ps -8 "device" }} {{ps -16 "alias"}} {{ps -8 "user"}} {{ps -22 "remote"}} {{ps -8 "network"}} {{ ps -6 "state"}}
{{- range . }}
{{ps -16 .UUID}} {{pt .AliveTime | ps -8}} {{ ps -8 .Device}} {{ps -16 .Alias}} {{ps -8 .User}} {{ps -22 .Remote}} {{ps -8 .Network}}  {{ ps -6 .State}}
{{- end }}
`
}

func (u VxLAN) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.VxLAN
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u VxLAN) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "vxlan",
		Aliases: []string{"vx"},
		Usage:   "VxLAN configuration",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all vxlan",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
