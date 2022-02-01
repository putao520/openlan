package v6

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/urfave/cli/v2"
)

type Network struct {
}

func (u Network) List(c *cli.Context) error {
	var listNo []NetworkDB
	if err := conf.OvS.List(doing, &listNo); err == nil {
		return api.Out(listNo, c.String("format"), "")
	}
	return nil
}

func (u Network) Add(c *cli.Context) error {
	name := c.String("name")
	address := c.String("address")
	newSw := NetworkDB{
		Name:    name,
		Address: address,
	}
	ops, err := conf.OvS.Create(&newSw)
	if err != nil {
		return err
	}
	libol.Debug("Network.Add %s", ops)
	if ret, err := conf.OvS.Transact(doing, ops...); err != nil {
		return err
	} else {
		libol.Debug("Network.Transact %s", ret)
	}
	return nil
}

func (u Network) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "network",
		Aliases: []string{"no"},
		Usage:   "Virtual network",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List network configuration",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add a virtual network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "name",
						Usage: "unique name with short long"},
					&cli.StringFlag{
						Name:  "address",
						Value: "169.254.169.0/24",
						Usage: "ip address for this network",
					},
				},
				Action: u.Add,
			},
		},
	})
}
