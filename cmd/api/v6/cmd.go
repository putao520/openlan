package v6

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/urfave/cli/v2"
)

func Before(c *cli.Context) error {
	if _, err := GetConf(); err == nil {
		return nil
	} else {
		return err
	}
}

func After(c *cli.Context) error {
	return nil
}

func Commands(app *api.App) {
	app.After = After
	app.Before = Before
	Switch{}.Commands(app)
	Network{}.Commands(app)
}
