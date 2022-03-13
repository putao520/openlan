package v6

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/danieldin95/openlan/pkg/database"
	"github.com/danieldin95/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
	"github.com/urfave/cli/v2"
	"sort"
	"strings"
)

type Link struct {
}

func (l Link) List(c *cli.Context) error {
	var lsLn []database.VirtualLink
	err := database.Client.List(&lsLn)
	if err != nil {
		return err
	}
	sort.SliceStable(lsLn, func(i, j int) bool {
		ii := lsLn[i]
		jj := lsLn[j]
		return ii.Connection > jj.Connection
	})
	return api.Out(lsLn, c.String("format"), "")
}

func GetUserPassword(auth string) (string, string) {
	auths := strings.SplitN(auth, ":", 2)
	if len(auths) == 2 {
		return auths[0], auths[1]
	}
	return auth, auth
}

func (l Link) Add(c *cli.Context) error {
	name := c.String("network")
	if name == "" {
		return libol.NewErr("network is nil")
	}
	lsVn := database.VirtualNetwork{Name: name}
	if err := database.Client.Get(&lsVn); err != nil {
		return libol.NewErr("find network %s: %s", name, err)
	}
	user, pass := GetUserPassword(c.String("auth"))
	newLn := database.VirtualLink{
		Network:    name,
		Connection: c.String("connection"),
		UUID:       database.GenUUID(),
		Device:     c.String("device"),
		Authentication: map[string]string{
			"username": user,
			"password": pass,
		},
		OtherConfig: map[string]string{
			"local_address":  lsVn.Address,
			"remote_address": c.String("address"),
		},
	}
	ops, err := database.Client.Create(&newLn)
	if err != nil {
		return err
	}
	libol.Debug("Link.Add %s %s", ops, lsVn)
	database.Client.Execute(ops)
	ops, err = database.Client.Where(&lsVn).Mutate(&lsVn, model.Mutation{
		Field:   &lsVn.LocalLinks,
		Mutator: ovsdb.MutateOperationInsert,
		Value:   []string{newLn.UUID},
	})
	if err != nil {
		return err
	}
	libol.Debug("Link.Add %s", ops)
	database.Client.Execute(ops)
	if ret, err := database.Client.Commit(); err != nil {
		return err
	} else {
		database.PrintError(ret)
	}
	return nil
}

func (l Link) Remove(c *cli.Context) error {
	name := c.String("network")
	lsVn := database.VirtualNetwork{Name: name}
	if err := database.Client.Get(&lsVn); err != nil {
		return libol.NewErr("find network %s: %s", name, err)
	}
	connection := c.String("connection")
	lsLn := database.VirtualLink{
		Network:    name,
		Connection: connection,
	}
	if err := database.Client.Get(&lsLn); err != nil {
		return libol.NewErr("find link %s: %s", connection, err)
	}
	ops, err := database.Client.Where(&lsLn).Delete()
	if err != nil {
		return err
	}
	libol.Debug("Link.Remove %s", ops)
	database.Client.Execute(ops)
	ops, err = database.Client.Where(&lsVn).Mutate(&lsVn, model.Mutation{
		Field:   &lsVn.LocalLinks,
		Mutator: ovsdb.MutateOperationDelete,
		Value:   []string{lsLn.UUID},
	})
	if err != nil {
		return err
	}
	libol.Debug("Link.Remove %s", ops)
	database.Client.Execute(ops)
	if ret, err := database.Client.Commit(); err != nil {
		return err
	} else {
		database.PrintError(ret)
	}
	return nil
}

func (l Link) Commands(app *api.App) {
	app.Command(&cli.Command{
		Name:    "link",
		Aliases: []string{"vl"},
		Usage:   "Virtual Link",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "List virtual links",
				Aliases: []string{"ls"},
				Action:  l.List,
			},
			{
				Name:  "add",
				Usage: "Add a virtual link",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "network",
						Usage: "the network name"},
					&cli.StringFlag{
						Name:  "connection",
						Usage: "connection for remote server",
					},
					&cli.StringFlag{
						Name:  "device",
						Usage: "the device name",
					},
					&cli.StringFlag{
						Name:  "auth",
						Usage: "user and password for authentication",
					},
					&cli.StringFlag{
						Name:  "address",
						Usage: "remote address in this link",
					},
				},
				Action: l.Add,
			},
			{
				Name:  "del",
				Usage: "Del a virtual link",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "network",
						Usage: "the network name"},
					&cli.StringFlag{
						Name:  "connection",
						Usage: "connection for remote server",
					},
				},
				Action: l.Remove,
			},
		},
	})
}
