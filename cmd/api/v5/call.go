package v5

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"strings"
)

func Before(c *cli.Context) error {
	token := c.String("token")
	if token == "" {
		tokenFile := api.AdminTokenFile
		if data, err := ioutil.ReadFile(tokenFile); err == nil {
			token = strings.TrimSpace(string(data))
		}
		_ = c.Set("token", token)
	}
	return nil
}

func After(c *cli.Context) error {
	return nil
}

func Call(app *api.App) {
	app.After = After
	app.Before = Before
	User{}.Commands(app)
	ACL{}.Commands(app)
	Device{}.Commands(app)
	Lease{}.Commands(app)
	Config{}.Commands(app)
	Point{}.Commands(app)
	VPNClient{}.Commands(app)
	Link{}.Commands(app)
	Server{}.Commands(app)
	Network{}.Commands(app)
	PProf{}.Commands(app)
	Esp{}.Commands(app)
	VxLAN{}.Commands(app)
	State{}.Commands(app)
	Policy{}.Commands(app)
}
