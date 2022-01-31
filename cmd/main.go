package main

import (
	"github.com/danieldin95/openlan/cmd/api"
	"github.com/danieldin95/openlan/cmd/api/v5"
	"github.com/danieldin95/openlan/cmd/api/v6"
	"log"
	"os"
)

func GetEnv(key, value string) string {
	val := os.Getenv(key)
	if val == "" {
		return value
	}
	return val
}

func main() {
	api.Version = GetEnv("OL_VERSION", api.Version)
	api.Url = GetEnv("OL_URL", api.Url)
	api.Token = GetEnv("OL_TOKEN", api.Token)
	api.Server = GetEnv("OL_CONF", api.Server)
	api.Database = GetEnv("OL_DATABASE", api.Database)
	app := &api.App{}
	app.New()

	switch api.Version {
	case "v6":
		v6.Switch{}.Commands(app)
	default:
		v5.User{}.Commands(app)
		v5.ACL{}.Commands(app)
		v5.Device{}.Commands(app)
		v5.Lease{}.Commands(app)
		v5.Config{}.Commands(app)
		v5.Point{}.Commands(app)
		v5.VPNClient{}.Commands(app)
		v5.Link{}.Commands(app)
		v5.Server{}.Commands(app)
		v5.Network{}.Commands(app)
		v5.PProf{}.Commands(app)
		v5.Esp{}.Commands(app)
		v5.VxLAN{}.Commands(app)
		v5.State{}.Commands(app)
		v5.Policy{}.Commands(app)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
