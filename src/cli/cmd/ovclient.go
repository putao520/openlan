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
{{ps -8 "uptime"}} {{ps -8 "name"}} {{ps -22 "remote"}} {{ ps -6 "state"}}
{{- range . }}
{{pi -8 .Uptime}} {{ps -8 .Name}} {{ps -22 .Remote}} {{ ps -6 .State}}
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
		Name:    "openvpn-client",
		Aliases: []string{"ovc"},
		Usage:   "Connected OpenVPN Client",
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
