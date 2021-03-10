package main

import (
	"fmt"
	"os"

	"github.com/loyalguru/go-kallax/generator/cli/kallax/cmd"

	"github.com/urfave/cli"
)

const version = "1.3.8"

func main() {
	if err := newApp().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
