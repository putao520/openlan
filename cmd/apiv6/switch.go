package apiv6

import (
	"context"
	"fmt"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/urfave/cli/v2"
)

type Switch struct {
}

func (u Switch) List(c *cli.Context) error {
	var lsList []GlobalSwitch
	if err := conf.OvS.List(context.Background(), &lsList); err == nil {
		for _, ls := range lsList {
			fmt.Printf("%+v\n", ls)
		}
	}
	return nil
}

func (u Switch) Add(c *cli.Context) error {
	protocol := c.String("protocol")
	listen := c.Int("listen")

	newSw := GlobalSwitch{
		Protocol: protocol,
		Listen:   listen,
	}

	var listSw []GlobalSwitch
	if err := conf.OvS.List(context.Background(), &listSw); err != nil {
		return err
	}
	if len(listSw) == 0 {
		ops, err := conf.OvS.Create(&newSw)
		if err != nil {
			return err
		}
		libol.Debug("Switch.Add %s", ops)
		if ret, err := conf.OvS.Transact(context.Background(), ops...); err != nil {
			return err
		} else {
			libol.Debug("OvS.Transact %s", ret)
		}
	} else {
		firstSw := GlobalSwitch{UUID: listSw[0].UUID}
		ops, err := conf.OvS.Where(&firstSw).Update(&newSw)
		if err != nil {
			return err
		}
		libol.Debug("Switch.Add %s", ops)
		if ret, err := conf.OvS.Transact(context.Background(), ops...); err != nil {
			return err
		} else {
			libol.Debug("OvS.Transact %s", ret)
		}
	}
	return nil
}

func (u Switch) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "switch",
		Aliases: []string{"sw"},
		Usage:   "Global switch",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display switch configuration",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add or Update switch",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "protocol", Value: "tcp"},
					&cli.IntFlag{Name: "listen", Value: 10002},
				},
				Action: u.Add,
			},
		},
	})
}
