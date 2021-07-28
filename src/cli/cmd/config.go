package cmd

import (
	"fmt"
	"github.com/danieldin95/openlan-go/src/config"
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/urfave/cli/v2"
	"path/filepath"
)

type Config struct {
	Cmd
}

func (u Config) Url(prefix, name string) string {
	return prefix + "/api/config"
}

func (u Config) List(c *cli.Context) error {
	url := u.Url(c.String("url"), "")
	url += "?format=" + c.String("format")
	clt := u.NewHttp(c.String("token"))
	if data, err := clt.GetBody(url); err == nil {
		fmt.Println(string(data))
		return nil
	} else {
		return err
	}
}

func (u Config) Check(c *cli.Context) error {
	out := u.Log()
	dir := c.String("dir")
	// Check proxy configurations.
	file := filepath.Join(dir, "proxy.json")
	if err := libol.FileExist(file); err == nil {
		obj := &config.Proxy{}
		if err := libol.UnmarshalLoad(obj, file); err != nil {
			out.Warn("%15s: %s", filepath.Base(file), err)
		} else {
			out.Info("%15s: %s", filepath.Base(file), "success")
		}
	}
	// Check OLAP configurations.
	file = filepath.Join(dir, "point.json")
	if err := libol.FileExist(file); err == nil {
		obj := &config.Point{}
		if err := libol.UnmarshalLoad(obj, file); err != nil {
			out.Warn("%15s: %s", filepath.Base(file), err)
		} else {
			out.Info("%15s: %s", filepath.Base(file), "success")
		}
	}
	// Check OLSW configurations.
	file = filepath.Join(dir, "switch", "switch.json")
	if err := libol.FileExist(file); err == nil {
		obj := &config.Switch{}
		if err := libol.UnmarshalLoad(obj, file); err != nil {
			out.Warn("%15s: %s", filepath.Base(file), err)
		} else {
			out.Info("%15s: %s", filepath.Base(file), "success")
		}
	}
	// Check network configurations.
	pattern := filepath.Join(dir, "switch", "network", "*.json")
	if files, err := filepath.Glob(pattern); err == nil {
		for _, file := range files {
			obj := &config.Network{}
			if err := libol.UnmarshalLoad(obj, file); err != nil {
				out.Warn("%15s: %s", filepath.Base(file), err)
			} else {
				out.Info("%15s: %s", filepath.Base(file), "success")
			}
		}
	}
	// Check ACL configurations.
	pattern = filepath.Join(dir, "switch", "acl", "*.json")
	if files, err := filepath.Glob(pattern); err == nil {
		for _, file := range files {
			obj := &config.ACL{}
			if err := libol.UnmarshalLoad(obj, file); err != nil {
				out.Warn("%15s: %s", filepath.Base(file), err)
			} else {
				out.Info("%15s: %s", filepath.Base(file), "success")
			}
		}
	}
	return nil
}

func (u Config) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "config",
		Aliases: []string{"cfg"},
		Usage:   "Switch configuration",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Usage:   "Display all configuration",
				Aliases: []string{"ls"},
				Action:  u.List,
			},
			{
				Name:    "check",
				Usage:   "Check all configuration",
				Aliases: []string{"co"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dir", Value: "/etc/openlan"},
				},
				Action: u.Check,
			},
		},
	})
}
