package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/src-d/go-kallax/generator"
)

type CmdGenerate struct {
	Input  string `short:"" long:"input" description:"input package directory" default:"."`
	Output string `short:"" long:"output" description:"output file name" default:"kallax.go"`
}

func (c *CmdGenerate) Execute(args []string) error {
	if !isDirectory(c.Input) {
		return fmt.Errorf("Input path should be a directory %s", c.Input)
	}

	p := generator.NewProcessor(c.Input, []string{c.Output})
	pkg, err := p.Do()
	if err != nil {
		return err
	}

	gen := generator.NewGenerator(filepath.Join(c.Input, c.Output))
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
