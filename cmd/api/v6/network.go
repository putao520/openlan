package v6

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/danieldin95/openlan/pkg/database"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
	"github.com/urfave/cli/v2"
	"sort"
)

type Network struct {
}

func (u Network) List(c *cli.Context) error {
	var listVn []database.VirtualNetwork
	err := database.Client.List(&listVn)
	if err != nil {
		return err
	}
	sort.SliceStable(listVn, func(i, j int) bool {
		ii := listVn[i]
		jj := listVn[j]
		return ii.Name > jj.Name
	})
	return api.Out(listVn, c.String("format"), "")
}

func (u Network) Add(c *cli.Context) error {
	name := c.String("name")
	if name == "" {
		return libol.NewErr("name is nil")
	}
	oldVn := database.VirtualNetwork{Name: name}
	if err := database.Client.Get(&oldVn); err == nil {
		return libol.NewErr("object existed with %s", oldVn.UUID)
	}
	address := c.String("address")
	newVn := database.VirtualNetwork{
		Name:    name,
		Address: address,
		Bridge:  "br-" + name,
		UUID:    database.GenUUID(),
	}
	ops, err := database.Client.Create(&newVn)
	if err != nil {
		return err
	}
	libol.Debug("Network.Add %s", ops)
	database.Client.Execute(ops)
	sw, err := database.Client.Switch()
	if err != nil {
		return err
	}
	ops, err = database.Client.Where(sw).Mutate(sw, model.Mutation{
		Field:   &sw.VirtualNetworks,
		Mutator: ovsdb.MutateOperationInsert,
		Value:   []string{newVn.UUID},
	})
	if err != nil {
		return err
	}
	libol.Debug("Network.Add %s", ops)
	database.Client.Execute(ops)
	if ret, err := database.Client.Commit(); err != nil {
		return err
	} else {
		database.PrintError(ret)
	}
	return nil
}

func (u Network) Remove(c *cli.Context) error {
	name := c.String("name")
	oldVn := database.VirtualNetwork{
		Name: name,
	}
	if err := database.Client.Get(&oldVn); err != nil {
		return err
	}
	ops, err := database.Client.Where(&oldVn).Delete()
	if err != nil {
		return err
	}
	libol.Debug("Switch.Remove %s", ops)
	database.Client.Execute(ops)
	sw, err := database.Client.Switch()
	if err != nil {
		return err
	}
	ops, err = database.Client.Where(sw).Mutate(sw, model.Mutation{
		Field:   &sw.VirtualNetworks,
		Mutator: ovsdb.MutateOperationDelete,
		Value:   []string{oldVn.UUID},
	})
	if err != nil {
		return err
	}
	libol.Debug("Network.Remove %s", ops)
	database.Client.Execute(ops)
	if ret, err := database.Client.Commit(); err != nil {
		return err
	} else {
		database.PrintError(ret)
	}
	return nil
}

func (u Network) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "network",
		Aliases: []string{"vn"},
		Usage:   "Virtual network",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List virtual networks",
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
			{
				Name:  "del",
				Usage: "Del a virtual network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "name",
						Usage: "unique name with short long"},
				},
				Action: u.Remove,
			},
		},
	})
}
