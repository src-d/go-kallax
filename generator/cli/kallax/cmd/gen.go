package cmd

import (
	"fmt"
	"path/filepath"

	"os"

	"gopkg.in/src-d/go-kallax.v1/generator"
	cli "gopkg.in/urfave/cli.v1"
)

var Generate = cli.Command{
	Name:   "gen",
	Usage:  "Generate kallax models",
	Action: generateAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "input",
			Value: ".",
			Usage: "Input package directory",
		},
		&cli.StringFlag{
			Name:  "output",
			Value: "kallax.go",
			Usage: "Output file name",
		},
		&cli.StringSliceFlag{
			Name:  "exclude, e",
			Usage: "List of excluded files from the package when generating the code for your models. Use this to exclude files in your package that uses the generated code. You can use this flag as many times as you want.",
		},
	},
}

func generateAction(c *cli.Context) error {
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

	var foundPrevious bool
	if _, err = os.Stat(output); err == nil {
		foundPrevious = true
		fmt.Fprintf(os.Stderr, "NOTE: Previous generated file `%s` found, renaming to `%s`\n", output, output+".old")
		err = os.Rename(output, output+".old")
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

	if foundPrevious {
		fmt.Fprintf(os.Stderr, "NOTE: Generation succeded, removing `%s`\n", output+".old")
		os.Remove(output + ".old")
	}

	return nil
}
