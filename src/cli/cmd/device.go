package cmd

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/urfave/cli/v2"
	"strings"
)

type Device struct {
	Cmd
}

func (u Device) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/device"
	} else {
		return prefix + "/api/device/" + name
	}
}

func (u Device) Tmpl() string {
	return strings.Join([]string{
		`# total {{ len . }}`,
		`{{ps -13 "name"}} {{ps -13 "mtu"}} {{ps -16 "mac"}} {{ps -6 "provider"}}`,
		`{{- range . }}`,
		`{{ps -13 .Name}} {{pi -13 .Mtu}} {{ps -16 .Mac}} {{ps -6 .Provider}}`,
		`{{- end }}`,
		``}, "\n")
}

func (u Device) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	client := Client{
		Auth: libol.Auth{
			Username: c.String("token"),
		},
	}
	request := client.NewRequest(url)
	var items []schema.Device
	if err := client.GetJSON(request, &items); err != nil {
		return err
	}
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u Device) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "device",
		Aliases: []string{"dev"},
		Usage:   "linux device",
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
