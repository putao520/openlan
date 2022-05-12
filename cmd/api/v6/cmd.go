package v6

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/danieldin95/openlan/pkg/database"
	"github.com/urfave/cli/v2"
)

func Before(c *cli.Context) error {
	if _, err := database.NewDBClient(nil); err == nil {
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
	Link{}.Commands(app)
	Name{}.Commands(app)
}
