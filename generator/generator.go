package generator

import (
	"fmt"
	"os"
)

// Generator is in charge of generating files for packages.
type Generator struct {
	filename string
}

// NewGenerator creates a new generator that can save on the given filename.
func NewGenerator(filename string) *Generator {
	return &Generator{filename}
}

// Generate writes the file with the contents of the given package.
func (g *Generator) Generate(pkg *Package) error {
	return g.writeFile(pkg)
}

func (g *Generator) writeFile(pkg *Package) (err error) {
	file, err := os.Create(g.filename)
	if err != nil {
		return err
	}

	defer func() {
		file.Close()
		if err != nil {
			fmt.Println()
			fmt.Println("kallax: No file generated due to an occurred error:")
			os.Remove(g.filename)
		}
	}()

	return Base.Execute(file, pkg)
}
