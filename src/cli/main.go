package main

import (
	"github.com/danieldin95/openlan-go/src/cli/cmd"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Usage:    "OpenLAN switch utility",
		Commands: []*cli.Command{},
	}

	cmd.User{}.Command(app)
	cmd.Switch{}.Command(app)
	cmd.ACL{}.Command(app)
	cmd.Link{}.Command(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
