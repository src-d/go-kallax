package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/src-d/go-kallax/generator"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "kallax"
	app.Version = "1.0.0"
	app.Usage = "generate kallax models"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "input",
			Value: ".",
			Usage: "Input package directory",
		},
		cli.StringFlag{
			Name:  "output",
			Value: "kallax.go",
			Usage: "Output file name",
		},
		cli.StringSliceFlag{
			Name:  "exclude, e",
			Usage: "List of excluded files in the excluded. You can use this flag as many times as you want.",
		},
	}
	app.Action = generateModels
	app.Commands = []cli.Command{
		{
			Name:   "gen",
			Usage:  "Generate kallax models",
			Action: app.Action,
			Flags:  app.Flags,
		},
	}

	app.Run(os.Args)
}

func generateModels(c *cli.Context) error {
	input := c.String("input")
	output := c.String("output")
	excluded := c.StringSlice("exclude")

	if !isDirectory(input) {
		return fmt.Errorf("kallax: Input path should be a directory %s", input)
	}

	p := generator.NewProcessor(input, excluded)
	pkg, err := p.Do()
	if err != nil {
		return err
	}

	gen := generator.NewGenerator(filepath.Join(input, output))
	err = gen.Generate(pkg)
	if err != nil {
		return err
	}

	return nil
}

func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}

	return info.IsDir()
}
