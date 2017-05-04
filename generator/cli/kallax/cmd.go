package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-kallax.v1/generator"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "kallax"
	app.Version = "1.1.3"
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
			Usage: "List of excluded files from the package when generating the code for your models. Use this to exclude files in your package that uses the generated code. You can use this flag as many times as you want.",
		},
	}
	app.Action = generateModels
	app.Commands = []cli.Command{
		{
			Name:   "gen",
			Usage:  "Generate kallax models",
			Action: generateModels,
			Flags:  app.Flags,
		},
		{
			Name:   "migrate",
			Usage:  "Generate migrations for current kallax models",
			Action: migrateAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "out, o",
					Usage: "Output directory of migrations",
				},
				cli.StringFlag{
					Name:  "name, n",
					Usage: "Descriptive name for the migration",
					Value: "migration",
				},
				cli.StringSliceFlag{
					Name:  "input, i",
					Usage: "List of directories to scan models from. You can use this flag as many times as you want.",
				},
			},
		},
	}

	app.Run(os.Args)
}

func generateModels(c *cli.Context) error {
	input := c.String("input")
	output := c.String("output")
	excluded := c.StringSlice("exclude")

	ok, err := isDirectory(input)
	if err != nil {
		return fmt.Errorf("kallax: can't check input directory: %s", err)
	}

	if !ok {
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

func migrateAction(c *cli.Context) error {
	dirs := c.StringSlice("input")
	dir := c.String("out")
	name := c.String("name")

	var pkgs []*generator.Package
	for _, dir := range dirs {
		ok, err := isDirectory(dir)
		if err != nil {
			return fmt.Errorf("kallax: cannot check directory in `input`: %s", err)
		}

		if !ok {
			return fmt.Errorf("kallax: `input` must be a valid directory")
		}

		p := generator.NewProcessor(dir, nil)
		p.Silent()
		pkg, err := p.Do()
		if err != nil {
			return err
		}

		pkgs = append(pkgs, pkg)
	}

	ok, err := isDirectory(dir)
	if err != nil {
		return fmt.Errorf("kallax: cannot check directory in `out`: %s", err)
	}

	if !ok {
		return fmt.Errorf("kallax: `out` must be a valid directory")
	}

	g := generator.NewMigrationGenerator(name, dir)
	migration, err := g.Build(pkgs...)
	if err != nil {
		return err
	}

	return g.Generate(migration)
}

func isDirectory(name string) (bool, error) {
	info, err := os.Stat(name)
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}
