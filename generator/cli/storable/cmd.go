package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewNamedParser("storable", flags.Default)
	parser.AddCommand(
		"gen",
		"Generate files for types using storable document.",
		"",
		&CmdGenerate{},
	)

	_, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrCommandRequired {
			parser.WriteHelp(os.Stdout)
		}

		os.Exit(1)
	}
}
