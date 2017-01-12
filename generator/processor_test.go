package generator

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type ProcessorSuite struct{}

var _ = Suite(&ProcessorSuite{})

func (s *ProcessorSuite) TestInit(c *C) {
	fixtureSrc := `
  package fixture

  import  "github.com/src-d/go-kallax"

  type InitExample struct {
    kallax.Document
    Foo string
  }
  
  func (i *InitExample) Init(doc kallax.DocumentBase) { }
  `

	pkg := s.processFixture(fixtureSrc)
	c.Assert(pkg.Models[0].Init, Equals, true)
}

func (s *ProcessorSuite) TestInitEmbedded(c *C) {
	fixtureSrc := `
  package fixture

  import  "github.com/src-d/go-kallax"

  type InitEmbeddedExample struct {
    kallax.Document
    OtherWithInit
  }

  type OtherWithInit struct {}

  func (i *OtherWithInit) Init(doc kallax.DocumentBase) error { return nil }
  `

	pkg := s.processFixture(fixtureSrc)
	c.Assert(pkg.Models, HasLen, 1)
	c.Assert(pkg.Models[0].Init, Equals, true)
}

func (s *ProcessorSuite) TestInlineStruct(c *C) {
	fixtureSrc := `
  package fixture

  import  "github.com/src-d/go-kallax"

  type Recur struct {
    kallax.Document
    Foo string
    R *Recur ` + "`bson:\",inline\"`" + `
  }
  `

	pkg := s.processFixture(fixtureSrc)
	c.Assert(pkg.Models[0].Fields[2].Fields[2].Inline(), Equals, true)
}

func (s *ProcessorSuite) TestTags(c *C) {
	fixtureSrc := `
	package fixture

	import 	"github.com/src-d/go-kallax"

	type Foo struct {
		kallax.Document
		Int int "foo"
	}
	`

	pkg := s.processFixture(fixtureSrc)
	c.Assert(pkg.Models[0].Fields[1].Tag, Equals, reflect.StructTag("foo"))
}

func (s *ProcessorSuite) TestRecursiveStruct(c *C) {
	fixtureSrc := `
	package fixture

	import 	"github.com/src-d/go-kallax"

	type Recur struct {
		kallax.Document
		Foo string
		R *Recur
	}
	`

	pkg := s.processFixture(fixtureSrc)

	c.Assert(
		pkg.Models[0].Fields[2].Fields[2].CheckedNode,
		Equals,
		pkg.Models[0].Fields[2].CheckedNode,
		Commentf("direct type recursivity not handled correctly."),
	)

	c.Assert(len(pkg.Models[0].Fields[2].Fields[2].Fields), Equals, 0)
}

func (s *ProcessorSuite) TestDeepRecursiveStruct(c *C) {
	fixtureSrc := `
	package fixture

	import 	"github.com/src-d/go-kallax"

	type Recur struct {
		kallax.Document
		Foo string
		R *Other
	}

	type Other struct {
		R *Recur
	}
	`

	pkg := s.processFixture(fixtureSrc)

	c.Assert(pkg.Models[0].Fields[2].Fields[0].Fields[2].CheckedNode, Equals, pkg.Models[0].Fields[2].CheckedNode, Commentf("direct type recursivity not handled correctly."))
	c.Assert(len(pkg.Models[0].Fields[2].Fields[0].Fields[2].Fields), Equals, 0)
}

func (s *ProcessorSuite) processFixture(source string) *Package {
	fset := &token.FileSet{}
	astFile, err := parser.ParseFile(fset, "fixture.go", source, 0)
	if err != nil {
		panic(err)
	}

	cfg := &types.Config{
		Importer: importer.For("gc", nil),
	}
	p, err := cfg.Check("foo", fset, []*ast.File{astFile}, nil)
	if err != nil {
		panic(err)
	}

	prc := NewProcessor("fixture", nil)
	prc.TypesPkg = p
	pkg, err := prc.processTypesPkg()
	if err != nil {
		panic(err)
	}

	return pkg
}
