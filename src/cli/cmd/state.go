package cmd

import (
	"github.com/danieldin95/openlan-go/src/schema"
	"github.com/urfave/cli/v2"
	"sort"
)

type State struct {
	Cmd
}

func (u State) Url(prefix, name string) string {
	if name == "" {
		return prefix + "/api/state"
	} else {
		return prefix + "/api/state/" + name
	}
}

func (u State) Tmpl() string {
	return `# total {{ len . }}
{{ps -16 "name"}} {{ps -8 "spi"}} {{ ps -16 "source" }} {{ ps -16 "destination" }} {{ ps -12 "bytes" }} {{ ps -12 "packages" }} 
{{- range . }}
{{ps -16 .Name}} {{pi -8 .Spi }} {{ ps -16 .Source }} {{ ps -16 .Dest }} {{ pi -12 .Bytes }} {{ pi -12 .Packages }}
{{- end }}
`
}

func (u State) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	clt := u.NewHttp(c.String("token"))
	var items []schema.EspState
	if err := clt.GetJSON(url, &items); err != nil {
		return err
	}
	sort.SliceStable(items, func(i, j int) bool {
		ii := items[i]
		jj := items[j]
		return ii.Spi > jj.Spi
	})
	return u.Out(items, c.String("format"), u.Tmpl())
}

func (u State) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "state",
		Aliases: []string{"se"},
		Usage:   "IPSec state configuration",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all xfrm state",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
		},
	})
}
