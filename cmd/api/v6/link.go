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

	if err := database.Client.List(&lsLn); err != nil {
		return err
	} else {
		sort.SliceStable(lsLn, func(i, j int) bool {
			ii := lsLn[i]
			jj := lsLn[j]
			return ii.Connection > jj.Connection
		})
		return api.Out(lsLn, c.String("format"), "")
	}
}

func GetUserPassword(auth string) (string, string) {
	auths := strings.SplitN(auth, ":", 2)
	if len(auths) == 2 {
		return auths[0], auths[1]
	}
	return auth, auth
}

func GetDeviceName(conn, provider, device string) string {
	if libol.GetPrefix(conn, 4) == "spi:" {
		return strings.Replace(conn, ":", "", 1)
	} else {
		if provider == "esp" {
			return "spi" + device
		}
		return "vir" + device
	}
}

func (l Link) Add(c *cli.Context) error {
	auth := c.String("authentication")
	connection := c.String("connection")
	lsLn := database.VirtualLink{
		UUID:       c.String("uuid"),
		Network:    c.String("network"),
		Connection: connection,
	}
	remoteAddr := c.String("remote-address")
	user, pass := GetUserPassword(auth)
	deviceId := c.String("device-id")
	if err := database.Client.Get(&lsLn); err == nil {
		lsVn := database.VirtualNetwork{
			Name: lsLn.Network,
		}
		if lsVn.Name == "" {
			return libol.NewErr("network is nil")
		}
		if err := database.Client.Get(&lsVn); err != nil {
			return libol.NewErr("find network %s: %s", lsVn.Name, err)
		}
		newLn := lsLn
		if connection != "" {
			newLn.Connection = connection
		}
		if user != "" {
			newLn.Authentication["username"] = user
		}
		if pass != "" {
			newLn.Authentication["password"] = pass
		}
		if remoteAddr != "" {
			newLn.OtherConfig["remote_address"] = remoteAddr
		}
		if deviceId != "" {
			newLn.Device = GetDeviceName(connection, lsVn.Provider, deviceId)
		}
		ops, err := database.Client.Where(&lsLn).Update(&newLn)
		if err != nil {
			return err
		}
		if ret, err := database.Client.Transact(ops...); err != nil {
			return err
		} else {
			database.PrintError(ret)
		}
	} else {
		lsVn := database.VirtualNetwork{
			Name: c.String("network"),
		}
		if lsVn.Name == "" {
			return libol.NewErr("network is nil")
		}
		if err := database.Client.Get(&lsVn); err != nil {
			return libol.NewErr("find network %s: %s", lsVn.Name, err)
		}
		uuid := c.String("uuid")
		if uuid == "" {
			uuid = database.GenUUID()
		}
		newLn := database.VirtualLink{
			Network:    lsLn.Network,
			Connection: lsLn.Connection,
			UUID:       uuid,
			Device:     GetDeviceName(connection, lsVn.Provider, deviceId),
			Authentication: map[string]string{
				"username": user,
				"password": pass,
			},
			OtherConfig: map[string]string{
				"local_address":  lsVn.Address,
				"remote_address": remoteAddr,
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
	}
	return nil
}

func (l Link) Remove(c *cli.Context) error {
	lsLn := database.VirtualLink{
		Network:    c.String("network"),
		Connection: c.String("connection"),
		UUID:       c.String("uuid"),
	}
	if err := database.Client.Get(&lsLn); err != nil {
		return err
	}
	lsVn := database.VirtualNetwork{
		Name: lsLn.Network,
	}
	if err := database.Client.Get(&lsVn); err != nil {
		return libol.NewErr("find network %s: %s", lsVn.Name, err)
	}
	if err := database.Client.Get(&lsLn); err != nil {
		return err
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
						Name: "uuid",
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "the network name",
					},
					&cli.StringFlag{
						Name:  "connection",
						Usage: "connection for remote server",
					},
					&cli.StringFlag{
						Name:  "device-id",
						Usage: "the device index",
					},
					&cli.StringFlag{
						Name:  "authentication",
						Usage: "user and password for authentication",
					},
					&cli.StringFlag{
						Name:  "remote-address",
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
						Name: "uuid",
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "the network name",
					},
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
