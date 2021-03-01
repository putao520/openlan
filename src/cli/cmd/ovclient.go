package cmd

import (
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/urfave/cli/v2"
)

type OvClient struct {
	Cmd
}

func (u OvClient) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/ovclient"
	} else {
		return prefix + "/api/ovclient/" + name
	}
}

func (u OvClient) Tmpl() string {
	return `# total {{ len . }}
{{ps -8 "alive"}} {{ps -16 "address"}} {{ps -15 "name"}} {{ps -22 "remote"}} {{ ps -6 "state"}}
{{- range . }}
{{pt .AliveTime | ps -8}} {{ps -16 .Address}} {{ps -15 .Name}} {{ps -22 .Remote}} {{ ps -6 .State}}
{{- end }}
`
}

func (u OvClient) List(c *cli.Context) error {
	url := u.Url(c.String("url"), c.String("network"))
	clt := u.NewHttp(c.String("token"))
	var items []schema.OvClient
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u OvClient) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "openvpn",
		Aliases: []string{"ov"},
		Usage:   "Connected client by OpenVPN",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all clients",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
