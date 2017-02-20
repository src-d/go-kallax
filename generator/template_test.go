package generator

import (
	"bytes"
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
return (*kallax.NumericID)(&r.ID), nil
case "foo":
return &r.Foo, nil
case "bar":
return types.Nullable(&r.Bar), nil
case "arr":
return types.Slice(&r.Arr), nil
case "arr_aliased":
return types.Slice(&r.ArrAliased), nil
case "urlarr":
return types.Slice((*[]*url.URL)(&r.URLArr)), nil
case "json":
return types.JSON(&r.JSON), nil
case "url":
return (*types.URL)(r.URL), nil
case "url_no_ptr":
return (*types.URL)(&r.UrlNoPtr), nil
case "foo_id":
return kallax.VirtualColumn("foo_id", r, new(kallax.NumericID)), nil
`

const baseTpl = `
	package fixture

	import "github.com/src-d/go-kallax"
	import "net/url"

	type Rel struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
	}

	type JSON struct {
		Foo string
	}

	type Baz string

	type URLs []*url.URL

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
		Bar *string
		Rel Rel
		Arr []string
		ArrAliased []Baz
		URLArr URLs
		JSON JSON
		URL *url.URL
		UrlNoPtr url.URL
		RelInverse Rel ` + "`fk:\",inverse\"`" + `
	}
`

func (s *TemplateSuite) TestGenColumnAddresses() {
	s.processSource(baseTpl)

	m := findModel(s.td.Package, "Foo")
	result := s.td.GenColumnAddresses(m)
	s.Equal(expectedAddresses, result)
}

const expectedValues = `case "id":
return r.ID, nil
case "foo":
return r.Foo, nil
case "bar":
if r.Bar == (*string)(nil) {
	return nil, nil
}
return r.Bar, nil
case "aliased":
return (string)(r.Aliased), nil
case "arr":
return types.Slice(r.Arr), nil
case "json":
return types.JSON(r.JSON), nil
case "url":
if r.URL == (*url.URL)(nil) {
	return nil, nil
}
return (*types.URL)(r.URL), nil
case "url_no_ptr":
return (*types.URL)(&r.UrlNoPtr), nil
case "foo_id":
return r.Model.VirtualColumn(col), nil
`

func (s *TemplateSuite) TestGenColumnValues() {
	s.processSource(`
	package fixture

	import "github.com/src-d/go-kallax"
	import "net/url"

	type Aliased string

	type Rel struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
	}

	type JSON struct {
		Foo string
	}

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
		Bar *string
		Rel Rel
		Aliased Aliased
		Arr []string
		JSON JSON
		URL *url.URL
		UrlNoPtr url.URL
		RelInverse Rel ` + "`fk:\",inverse\"`" + `
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
kallax.NewSchemaField("arr_aliased"),
kallax.NewSchemaField("urlarr"),
kallax.NewSchemaField("json"),
kallax.NewSchemaField("url"),
kallax.NewSchemaField("url_no_ptr"),
kallax.NewSchemaField("foo_id"),
`

func (s *TemplateSuite) TestGenModelColumns() {
	s.processSource(baseTpl)
	m := findModel(s.td.Package, "Foo")
	result := s.td.GenModelColumns(m)
	s.Equal(expectedColumns, result)
}

const jsonBaseTpl = `
	package fixture

	import "github.com/src-d/go-kallax"
	import "net/url"

	type Rel struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
	}

	type Deep struct {
		X int ` + "`json:\"redefined\"`" + `
		Y int
	}

	type Bar struct {
		A bool
	}

	type JSON struct {
		Foo string
		Other Bar
		Arr []Deep
	}

	type JS struct {
		Foo string
	}

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
		Bar *string
		Arr []string
		JSON JSON
		URL *url.URL
		UrlNoPtr url.URL
		Rel Rel
		JSONArray []JS
	}
`

const expectedSchema = `ID kallax.SchemaField
Foo kallax.SchemaField
Bar kallax.SchemaField
Arr kallax.SchemaField
JSON *schemaFooJSON
URL kallax.SchemaField
UrlNoPtr kallax.SchemaField
JSONArray *schemaFooJSONArray
`

const expectedSubSchemas = `type schemaFooJSON struct {
*kallax.BaseSchemaField
Foo kallax.SchemaField
Other *schemaFooJSONOther
Arr *schemaFooJSONArr
}

type schemaFooJSONArr struct {
*kallax.JSONSchemaArray
X kallax.SchemaField
Y kallax.SchemaField
}

func (s *schemaFooJSONArr) At(n int) *schemaFooJSONArr {
return &schemaFooJSONArr{
JSONSchemaArray: kallax.NewJSONSchemaArray("json", "Arr"),
X:kallax.NewJSONSchemaKey(kallax.JSONInt, "json", "Arr", fmt.Sprint(n), "redefined"),
Y:kallax.NewJSONSchemaKey(kallax.JSONInt, "json", "Arr", fmt.Sprint(n), "Y"),
}
}

type schemaFooJSONArray struct {
*kallax.BaseSchemaField
Foo kallax.SchemaField
}

func (s *schemaFooJSONArray) At(n int) *schemaFooJSONArray {
return &schemaFooJSONArray{
BaseSchemaField: kallax.NewSchemaField("jsonarray").(*kallax.BaseSchemaField),
Foo:kallax.NewJSONSchemaKey(kallax.JSONText, "jsonarray", fmt.Sprint(n), "Foo"),
}
}

type schemaFooJSONOther struct {
*kallax.BaseSchemaField
A kallax.SchemaField
}

`

func (s *TemplateSuite) TestGenModelSchema() {
	s.processSource(jsonBaseTpl)
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
Foo:kallax.NewJSONSchemaKey(kallax.JSONText, "json", "Foo"),
Other:&schemaFooJSONOther{
JSONSchemaKey: kallax.NewJSONSchemaKey(kallax.JSONAny, "json", "Other"),
A:kallax.NewJSONSchemaKey(kallax.JSONBool, "json", "Other", "A"),
},
Arr:&schemaFooJSONArr{
JSONSchemaArray: kallax.NewJSONSchemaArray("json", "Arr"),
X:kallax.NewJSONSchemaKey(kallax.JSONInt, "json", "Arr", "redefined"),
Y:kallax.NewJSONSchemaKey(kallax.JSONInt, "json", "Arr", "Y"),
},
},
URL:kallax.NewSchemaField("url"),
UrlNoPtr:kallax.NewSchemaField("url_no_ptr"),
JSONArray:&schemaFooJSONArray{
BaseSchemaField: kallax.NewSchemaField("jsonarray").(*kallax.BaseSchemaField),
Foo:kallax.NewJSONSchemaKey(kallax.JSONText, "jsonarray", "Foo"),
},
`

func (s *TemplateSuite) TestGenSchemaInit() {
	s.processSource(jsonBaseTpl)
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
		ID int64 ` + "`pk:\"autoincr\"`" + `
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
		ID int64 ` + "`pk:\"autoincr\"`" + `
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
	s.Nil(err)
}

func TestTemplate(t *testing.T) {
	suite.Run(t, new(TemplateSuite))
}
