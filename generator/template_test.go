package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TemplateSuite struct {
	suite.Suite
	td *TemplateData
}

func (s *TemplateSuite) SetupTest() {
	s.td = &TemplateData{
		nil,
		make(map[interface{}]string),
		make(map[string]*Field),
	}
}

func (s *TemplateSuite) processSource(source string) {
	fset := &token.FileSet{}
	astFile, err := parser.ParseFile(fset, "fixture.go", source, 0)
	s.Nil(err)

	cfg := &types.Config{
		Importer: importer.For("gc", nil),
	}
	p, err := cfg.Check("foo", fset, []*ast.File{astFile}, nil)
	s.Nil(err)

	prc := NewProcessor("fixture", []string{"foo.go"})
	prc.Package = p
	s.td.Package, err = prc.processPackage()
	s.Nil(err)
}

const expectedAddresses = `case "id":
return &r.Model.ID, nil
case "foo":
return &r.Foo, nil
case "bar":
return r.Bar, nil
case "arr":
return types.Slice(&r.Arr), nil
case "json":
return types.JSON(&r.JSON), nil
case "url":
return (*types.URL)(r.URL), nil
case "url_no_ptr":
return (*types.URL)(&r.UrlNoPtr), nil
`

const baseTpl = `
	package fixture

	import "github.com/src-d/go-kallax"
	import "net/url"

	type Rel struct {
		kallax.Model
		Foo string
	}

	type JSON struct {
		Foo string
	}

	type Foo struct {
		kallax.Model
		Foo string
		Bar *string
		Rel Rel
		Arr []string
		JSON JSON
		URL *url.URL
		UrlNoPtr url.URL
	}
`

func (s *TemplateSuite) TestGenColumnAddresses() {
	s.processSource(baseTpl)

	m := findModel(s.td.Package, "Foo")
	result := s.td.GenColumnAddresses(m)
	s.Equal(expectedAddresses, result)
}

const expectedValues = `case "id":
return r.Model.ID, nil
case "foo":
return r.Foo, nil
case "bar":
return r.Bar, nil
case "aliased":
return (string)(r.Aliased), nil
case "arr":
return types.Slice(r.Arr), nil
case "json":
return types.JSON(r.JSON), nil
case "url":
return (*types.URL)(r.URL), nil
case "url_no_ptr":
return (*types.URL)(&r.UrlNoPtr), nil
`

func (s *TemplateSuite) TestGenColumnValues() {
	s.processSource(`
	package fixture

	import "github.com/src-d/go-kallax"
	import "net/url"

	type Aliased string

	type Rel struct {
		kallax.Model
		Foo string
	}

	type JSON struct {
		Foo string
	}

	type Foo struct {
		kallax.Model
		Foo string
		Bar *string
		Rel Rel
		Aliased Aliased
		Arr []string
		JSON JSON
		URL *url.URL
		UrlNoPtr url.URL
	}
	`)

	m := findModel(s.td.Package, "Foo")
	result := s.td.GenColumnValues(m)
	s.Equal(expectedValues, result)
}

const expectedColumns = `kallax.NewSchemaField("id"),
kallax.NewSchemaField("foo"),
kallax.NewSchemaField("bar"),
kallax.NewSchemaField("arr"),
kallax.NewSchemaField("json"),
kallax.NewSchemaField("url"),
kallax.NewSchemaField("url_no_ptr"),
`

func (s *TemplateSuite) TestGenModelColumns() {
	s.processSource(baseTpl)
	m := findModel(s.td.Package, "Foo")
	result := s.td.GenModelColumns(m)
	s.Equal(expectedColumns, result)
}

const expectedSchema = `ID kallax.SchemaField
Foo kallax.SchemaField
Bar kallax.SchemaField
Arr kallax.SchemaField
JSON *schemaFooJSON
URL kallax.SchemaField
UrlNoPtr kallax.SchemaField
`

const expectedSubSchemas = `type schemaFooJSON struct {
*kallax.BaseSchemaField
Foo kallax.SchemaField
}

`

func (s *TemplateSuite) TestGenModelSchema() {
	s.processSource(baseTpl)
	m := findModel(s.td.Package, "Foo")
	result := s.td.GenModelSchema(m)
	s.Equal(expectedSchema, result)
	s.Equal(expectedSubSchemas, s.td.GenSubSchemas())
}

const expectedInit = `ID:kallax.NewSchemaField("id"),
Foo:kallax.NewSchemaField("foo"),
Bar:kallax.NewSchemaField("bar"),
Arr:kallax.NewSchemaField("arr"),
JSON:&schemaFooJSON{
BaseSchemaField: kallax.NewSchemaField("json").(*kallax.BaseSchemaField),
Foo:kallax.NewSchemaField("Foo"),
},
URL:kallax.NewSchemaField("url"),
UrlNoPtr:kallax.NewSchemaField("url_no_ptr"),
`

func (s *TemplateSuite) TestGenSchemaInit() {
	s.processSource(baseTpl)
	m := findModel(s.td.Package, "Foo")

	s.Equal(expectedInit, s.td.GenSchemaInit(m))
}

func (s *TemplateSuite) TestGenTypeName() {
	s.processSource(`
	package fixture

	import "github.com/src-d/go-kallax"
	import "net/url"

	type Foo struct {
		kallax.Model
		Slice []string
		Ptr *string
		NoPtr string
		URL *url.URL
		UrlNoPtr url.URL
	}
	`)

	m := findModel(s.td.Package, "Foo")
	var cases = []struct {
		field    string
		expected string
	}{
		{"Slice", "string"},
		{"Ptr", "string"},
		{"NoPtr", "string"},
		{"URL", "url.URL"},
		{"UrlNoPtr", "url.URL"},
	}

	for _, c := range cases {
		s.Equal(c.expected, s.td.GenTypeName(findField(m, c.field)), c.field)
	}
}

func (s *TemplateSuite) TestIsPtrSlice() {
	s.processSource(`
	package fixture

	import "github.com/src-d/go-kallax"
	import "net/url"

	type Foo struct {
		kallax.Model
		Ptr *url.URL
		Slice []url.URL
		PtrSlice []*url.URL
	}
	`)

	m := findModel(s.td.Package, "Foo")
	var cases = []struct {
		field    string
		expected bool
	}{
		{"Ptr", false},
		{"Slice", false},
		{"PtrSlice", true},
	}

	for _, c := range cases {
		s.Equal(c.expected, s.td.IsPtrSlice(findField(m, c.field)), c.field)
	}
}

func (s *TemplateSuite) TestExecute() {
	s.processSource(baseTpl)
	var buf bytes.Buffer
	err := Base.Execute(&buf, s.td.Package)
	fmt.Println(err)
	s.Nil(err)
}

func TestTemplate(t *testing.T) {
	suite.Run(t, new(TemplateSuite))
}
