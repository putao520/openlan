package v6

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/danieldin95/openlan/pkg/database"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/urfave/cli/v2"
	"net"
	"sort"
)

type Name struct {
}

func (u Name) List(c *cli.Context) error {
	var listNa []database.NameCache
	if err := database.Client.List(&listNa); err != nil {
		return err
	} else {
		sort.SliceStable(listNa, func(i, j int) bool {
			ii := listNa[i]
			jj := listNa[j]
			return ii.Name > jj.Name
		})
		return api.Out(listNa, c.String("format"), "")
	}
}

func (u Name) Add(c *cli.Context) error {
	lsNa := database.NameCache{
		Name: c.String("name"),
	}
	if lsNa.Name == "" {
		return libol.NewErr("Name is nil")
	}
	address := c.String("address")
	if address == "" {
		addrIps, _ := net.LookupIP(lsNa.Name)
		if len(addrIps) > 0 {
			address = addrIps[0].String()
		}
	}
	newNa := database.NameCache{
		Name:    lsNa.Name,
		Address: address,
	}
	if err := database.Client.Get(&lsNa); err == nil {
		if len(address) > 0 && lsNa.Address != address {
			ops, err := database.Client.Where(&lsNa).Update(&newNa)
			if err != nil {
				return err
			}
			if ret, err := database.Client.Transact(ops...); err != nil {
				return err
			} else {
				database.PrintError(ret)
			}
		}
	} else {
		ops, err := database.Client.Create(&newNa)
		if err != nil {
			return err
		}
		libol.Debug("Name.Add %s", ops)
		if ret, err := database.Client.Transact(ops...); err != nil {
			return err
		} else {
			database.PrintError(ret)
		}
	}
	return nil
}

func (u Name) Remove(c *cli.Context) error {
	lsNa := database.NameCache{
		Name: c.String("name"),
	}
	if err := database.Client.Get(&lsNa); err != nil {
		return nil
	}
	ops, err := database.Client.Where(&lsNa).Delete()
	if err != nil {
		return err
	}
	libol.Debug("Name.Remove %s", ops)
	database.Client.Execute(ops)
	if ret, err := database.Client.Commit(); err != nil {
		return err
	} else {
		database.PrintError(ret)
	}
	return nil
}

func (u Name) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "name",
		Aliases: []string{"na"},
		Usage:   "Name cache",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List name cache",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:  "add",
				Usage: "Add or update name cache",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "name",
					},
					&cli.StringFlag{
						Name: "address",
					},
				},
				Action: u.Add,
			},
			{
				Name:  "del",
				Usage: "Delete a name cache",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "name",
					},
				},
				Action: u.Remove,
			},
		},
	})
}
