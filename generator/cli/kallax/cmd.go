package main

import (
	"os"

	"gopkg.in/src-d/go-kallax.v1/generator/cli/kallax/cmd"

	"gopkg.in/urfave/cli.v1"
)

const version = "1.2.7"

func main() {
	newApp().Run(os.Args)
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "kallax"
	app.Version = version
	app.Usage = "generate kallax models"
	app.Flags = cmd.Generate.Flags
	app.Action = cmd.Generate.Action
	app.Commands = cli.Commands{
		cmd.Generate,
		cmd.Migrate,
	}

	return app
}
