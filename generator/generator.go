package generator

import "os"

type Generator struct {
	filename string
}

func NewGenerator(filename string) *Generator {
	return &Generator{filename}
}

func (g *Generator) Generate(pkg *Package) error {
	return g.writeFile(pkg)
}

func (g *Generator) writeFile(pkg *Package) error {
	file, err := os.Create(g.filename)
	if err != nil {
		return err
	}

	defer file.Close()
	return Base.Execute(file, pkg)
}
