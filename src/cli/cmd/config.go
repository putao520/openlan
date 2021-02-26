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
	dir := c.String("dir")
	// Check proxy configurations.
	pxFile := filepath.Join(dir, "proxy.json")
	if err := libol.FileExist(pxFile); err == nil {
		obj := &config.Proxy{}
		if err := libol.UnmarshalLoad(obj, pxFile); err != nil {
			libol.Warn("proxy.json: %s", err)
		}
	}
	// Check OLAP configurations.
	ptFile := filepath.Join(dir, "point.json")
	if err := libol.FileExist(ptFile); err == nil {
		obj := &config.Point{}
		if err := libol.UnmarshalLoad(obj, ptFile); err != nil {
			libol.Warn("point.json: %s", err)
		}
	}
	// Check OLSW configurations.
	swFile := filepath.Join(dir, "switch", "switch.json")
	if err := libol.FileExist(swFile); err == nil {
		obj := &config.Switch{}
		if err := libol.UnmarshalLoad(obj, swFile); err != nil {
			libol.Warn("switch.json: %s", err)
		}
	}
	// Check network configurations.
	pattern := filepath.Join(dir, "switch", "network", "*.json")
	if files, err := filepath.Glob(pattern); err == nil {
		for _, file := range files {
			obj := &config.Network{}
			if err := libol.UnmarshalLoad(obj, file); err != nil {
				libol.Warn("%s: %s", filepath.Base(file), err)
			}
		}
	}
	return nil
}

func (u Config) Commands(app *cli.App) cli.Commands {
	return append(app.Commands, &cli.Command{
		Name:    "config",
		Aliases: []string{"cfg"},
		Usage:   "Configuration",
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
