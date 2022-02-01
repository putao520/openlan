package v6

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/urfave/cli/v2"
)

type Switch struct {
}

func (u Switch) List(c *cli.Context) error {
	var listSw []SwitchDB
	if err := conf.OvS.List(doing, &listSw); err == nil {
		return api.Out(listSw, c.String("format"), "")
	}
	return nil
}

func (u Switch) Add(c *cli.Context) error {
	protocol := c.String("protocol")
	listen := c.Int("listen")
	var listSw []SwitchDB
	if err := conf.OvS.List(doing, &listSw); err != nil {
		return err
	}
	newSw := SwitchDB{
		Protocol: protocol,
		Listen:   listen,
	}
	if len(listSw) == 0 {
		ops, err := conf.OvS.Create(&newSw)
		if err != nil {
			return err
		}
		libol.Debug("Switch.Add %s", ops)
		if ret, err := conf.OvS.Transact(doing, ops...); err != nil {
			return err
		} else {
			libol.Debug("Switch.Transact %s", ret)
		}
	} else {
		ops, err := conf.OvS.Where(&listSw[0]).Update(&newSw)
		if err != nil {
			return err
		}
		libol.Debug("Switch.Add %s", ops)
		if ret, err := conf.OvS.Transact(doing, ops...); err != nil {
			return err
		} else {
			libol.Debug("Switch.Add %s", ret)
		}
	}
	return nil
}

func (u Switch) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "switch",
		Aliases: []string{"sw"},
		Usage:   "Global switch",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List switch configuration",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add or update a switch",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "protocol",
						Value: "tcp",
						Usage: "used protocol: tcp|udp|http|tls"},
					&cli.IntFlag{
						Name:  "listen",
						Value: 10002,
						Usage: "listen on port: 1024-65535",
					},
				},
				Action: u.Add,
			},
		},
	})
}
