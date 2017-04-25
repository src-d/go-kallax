package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/src-d/go-parse-utils.v1"
)

type ProcessorSuite struct {
	suite.Suite
}

func (s *ProcessorSuite) TestInlineStruct() {
	fixtureSrc := `
  package fixture

  import  "gopkg.in/src-d/go-kallax.v1"

  type Foo struct {}

  type Bar struct {
    kallax.Model
	ID int64 ` + "`pk:\"autoincr\"`" + `
    Foo string
    R *Foo ` + "`kallax:\",inline\"`" + `
  }
  `

	pkg := s.processFixture(fixtureSrc)
	s.True(findModel(pkg, "Bar").Fields[3].Inline())
}

func (s *ProcessorSuite) TestTags() {
	fixtureSrc := `
	package fixture

	import 	"gopkg.in/src-d/go-kallax.v1"

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Int int "foo"
	}
	`

	pkg := s.processFixture(fixtureSrc)
	s.Equal(reflect.StructTag("foo"), findModel(pkg, "Foo").Fields[2].Tag)
}

func (s *ProcessorSuite) TestRecursiveModel() {
	fixtureSrc := `
	package fixture

	import 	"gopkg.in/src-d/go-kallax.v1"

	type Recur struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
		R *Recur
	}
	`

	pkg := s.processFixture(fixtureSrc)
	m := findModel(pkg, "Recur")

	s.Equal(findField(m, "R").Kind, Relationship)
	s.Len(findField(m, "R").Fields, 0)
}

func (s *ProcessorSuite) TestDeepRecursiveStruct() {
	fixtureSrc := `
	package fixture

	import 	"gopkg.in/src-d/go-kallax.v1"

	type Recur struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
		Rec *Other
	}

	type Other struct {
		R *Recur
	}
	`

	pkg := s.processFixture(fixtureSrc)
	m := findModel(pkg, "Recur")

	s.Equal(
		m.Fields[3].Fields[0].Fields[3].Node,
		m.Fields[3].Node,
		"indirect type recursivity not handled correctly.",
	)
	s.Len(pkg.Models[0].Fields[3].Fields[0].Fields[3].Fields, 0)
}

func (s *ProcessorSuite) TestIsEventPresent() {
	fixtureSrc := `
	package fixture

	import 	"gopkg.in/src-d/go-kallax.v1"

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
	}

	func (r *Foo) BeforeUpdate() error {
		return nil
	}

	func (r *Foo) BeforeInsert() int {
		return 0
	}

	func (r *Foo) AfterInsert() int {
		return 0
	}

	func (r *Foo) AfterUpdate(foo int) {
	}
	`

	p := s.processorFixture(fixtureSrc)
	pkg, err := p.processPackage()
	s.Nil(err)

	m := findModel(pkg, "Foo")
	s.True(p.isEventPresent(m.Node, BeforeUpdate))
	s.False(p.isEventPresent(m.Node, BeforeInsert))
	s.False(p.isEventPresent(m.Node, AfterInsert))
	s.False(p.isEventPresent(m.Node, AfterUpdate))
}

func (s *ProcessorSuite) TestProcessField() {
	fixtureSrc := `
	package fixture

	import 	"gopkg.in/src-d/go-kallax.v1"
	import "database/sql/driver"

	type BasicAlias string
	type MapAlias map[string]int
	type SliceAlias []string
	type ArrayAlias [4]string

	type Related struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
	}

	type JSON struct {
		Bar string
	}

	type Interface interface {
		Foo()
	}

	type SQLInterface interface {
		Scan(interface{}) error
		Value(interface{}) (driver.Value, error)
	}

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Basic string
		AliasBasic BasicAlias
		BasicPtr *string
		Relationship Related
		RelSlice []Related
		RelArray [4]Related
		Map map[string]interface{}
		MapAlias MapAlias
		AliasSlice SliceAlias
		BasicSlice []string
		ComplexSlice []JSON
		JSON JSON
		JSONPtr *JSON
		AliasArray ArrayAlias
		BasicArray [4]string
		ComplexArray [4]JSON
		InlineArray struct{A int}
		Interface Interface
		SQLInterface SQLInterface
	}
	`

	pkg := s.processFixture(fixtureSrc)
	cases := []struct {
		name    string
		kind    FieldKind
		isJSON  bool
		isAlias bool
		isPtr   bool
	}{
		{"Basic", Basic, false, false, false},
		{"AliasBasic", Basic, false, true, false},
		{"BasicPtr", Basic, false, false, true},
		{"Relationship", Relationship, false, false, false},
		{"RelSlice", Relationship, false, false, false},
		{"Map", Map, true, false, false},
		{"MapAlias", Map, true, false, false},
		{"AliasSlice", Slice, false, true, false},
		{"BasicSlice", Slice, false, false, false},
		{"ComplexSlice", Slice, true, false, false},
		{"JSON", Struct, true, false, false},
		{"JSONPtr", Struct, true, false, true},
		{"AliasArray", Array, false, true, false},
		{"BasicArray", Array, false, false, false},
		{"ComplexArray", Array, true, false, false},
		{"InlineArray", Struct, true, false, false},
		{"Interface", Interface, true, false, false},
		{"SQLInterface", Interface, true, false, false}, // TODO false, false, false
	}

	m := findModel(pkg, "Foo")
	for _, c := range cases {
		f := findField(m, c.name)
		s.NotNil(f, "%s should not be nil", c.name)

		s.Equal(c.kind, f.Kind, "%s kind", c.name)
		s.Equal(c.isJSON, f.IsJSON, "%s is json", c.name)
		s.Equal(c.isAlias, f.IsAlias, "%s is alias", c.name)
		s.Equal(c.isPtr, f.IsPtr, "%s is ptr", c.name)
	}

	s.Nil(findField(m, "RelArray"), "RelArray should not be generated")
}

func (s *ProcessorSuite) TestCtor() {
	fixtureSrc := `
	package fixture

	import 	"gopkg.in/src-d/go-kallax.v1"

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
	}

	func newFoo() *Foo {
		return &Foo{}
	}
	`

	pkg := s.processFixture(fixtureSrc)
	m := findModel(pkg, "Foo")

	s.NotNil(m.CtorFunc, "Foo should have ctor")
}

func (s *ProcessorSuite) TestSQLTypeIsInterface() {
	fixtureSrc := `
	package fixture

	import "gopkg.in/src-d/go-kallax.v1"
	import "database/sql/driver"

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo Bar
	}

	type Bar string

	func (*Bar) Scan(v interface{}) error {
		return nil
	}

	func (Bar) Value() (driver.Value, error) {
		return nil, nil
	}
	`

	pkg := s.processFixture(fixtureSrc)
	field := findField(findModel(pkg, "Foo"), "Foo")
	s.Equal(Interface, field.Kind)
}

func (s *ProcessorSuite) TestIsSQLType() {
	fixtureSrc := `
	package fixture

	import 	"gopkg.in/src-d/go-kallax.v1"

	type SQLTypeFixture struct {
		kallax.Model
		ID kallax.ULID ` + "`pk:\"\"`" + `
		Foo string
	}
	`

	p := s.processorFixture(fixtureSrc)
	pkg, err := p.processPackage()
	s.Nil(err)
	m := findModel(pkg, "SQLTypeFixture")

	s.True(isSQLType(p.Package, types.NewPointer(m.ID.Node.Type())))
	// model is index 1 because ID is always index 0
	s.False(isSQLType(p.Package, types.NewPointer(m.Fields[1].Node.Type())))
}

func (s *ProcessorSuite) processorFixture(source string) *Processor {
	fset := &token.FileSet{}
	astFile, err := parser.ParseFile(fset, "fixture.go", source, 0)
	s.Nil(err)

	cfg := &types.Config{
		Importer: parseutil.NewImporter(),
	}
	p, err := cfg.Check("foo", fset, []*ast.File{astFile}, nil)
	s.Nil(err)

	prc := NewProcessor("fixture", []string{"foo.go"})
	prc.Package = p
	return prc
}

func (s *ProcessorSuite) processFixture(source string) *Package {
	pkg, err := s.processorFixture(source).processPackage()
	s.Nil(err)
	return pkg
}

func (s *ProcessorSuite) TestDo() {
	srcDir, err := getPkgSrcDir("gopkg.in/src-d/go-kallax.v1")
	s.Nil(err)
	p := NewProcessor(srcDir, []string{"README.md"})
	pkg, err := p.Do()
	s.NotNil(pkg)
	s.NoError(err)
}

func (s *ProcessorSuite) TestIsModel() {
	src := `
	package fixture

	import "gopkg.in/src-d/go-kallax.v1"

	type Bar struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Bar string
	}

	type Struct struct {
		Bar Bar
	}

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
		Ptr *Bar
		NoPtr Bar
		Struct Struct
	}	
	`
	pkg := s.processFixture(src)
	m := findModel(pkg, "Foo")
	cases := []struct {
		field    string
		expected bool
	}{
		{"Foo", false},
		{"Ptr", true},
		{"NoPtr", true},
		{"Struct", false},
	}

	for _, c := range cases {
		s.Equal(c.expected, isModel(findField(m, c.field).Node.Type(), true), c.field)
	}
}

func (s *ProcessorSuite) TestIsEmbedded() {
	src := `
	package fixture

	import "gopkg.in/src-d/go-kallax.v1"

	type Bar struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Bar string
	}

	type Struct struct {
		Bar Bar
	}

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		A Bar
		B *Bar
		Bar
		*Struct
		C struct {
			D int
		}
	}	
	`
	pkg := s.processFixture(src)
	m := findModel(pkg, "Foo")
	cases := []struct {
		field    string
		embedded bool
	}{
		{"Model", true},
		{"A", false},
		{"B", false},
		{"Bar", true},
		{"Struct", true},
		{"C", false},
	}

	for _, c := range cases {
		s.Equal(c.embedded, findField(m, c.field).IsEmbedded, c.field)
	}
}

func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorSuite))
}

func findModel(pkg *Package, name string) *Model {
	for _, m := range pkg.Models {
		if m.Name == name {
			return m
		}
	}
	return nil
}

func findField(m *Model, name string) *Field {
	for _, f := range m.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}
