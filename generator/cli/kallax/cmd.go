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
			Usage: "input package directory",
		},
		cli.StringFlag{
			Name:  "output",
			Value: "kallax.go",
			Usage: "output file name",
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

	if !isDirectory(input) {
		return fmt.Errorf("kallax: Input path should be a directory %s", input)
	}

	p := generator.NewProcessor(input, nil)
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
