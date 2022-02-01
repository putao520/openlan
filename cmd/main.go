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
		v6.Commands(app)
	default:
		v5.Commands(app)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
