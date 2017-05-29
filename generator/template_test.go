package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
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
	s.NoError(err)

	cfg := &types.Config{
		Importer: importer.For("gc", nil),
	}
	p, err := cfg.Check("foo", fset, []*ast.File{astFile}, nil)
	s.NoError(err)

	prc := NewProcessor("fixture", []string{"foo.go"})
	prc.Package = p
	s.td.Package, err = prc.processPackage()
	s.NoError(err)
}

const expectedAddresses = `case "id":
return (*kallax.NumericID)(&r.ID), nil
case "foo":
return &r.Foo, nil
case "bar":
return types.Nullable(&r.Bar), nil
case "baz":
return &r.Baz, nil
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
case "rel_id":
return types.Nullable(kallax.VirtualColumn("rel_id", r, new(kallax.NumericID))), nil
case "basic_alias":
return (*int)(&r.BasicAlias), nil
`

const baseTpl = `
	package fixture

	import "gopkg.in/src-d/go-kallax.v1"
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

	type Qux int

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		Foo string
		Bar *string
		Baz int64
		Rel Rel
		Arr []string
		ArrAliased []Baz
		URLArr URLs
		JSON JSON
		URL *url.URL
		UrlNoPtr url.URL
		RelInverse Rel ` + "`fk:\",inverse\"`" + `
		BasicAlias Qux
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
case "rel_id":
return r.Model.VirtualColumn(col), nil
`

func (s *TemplateSuite) TestGenColumnValues() {
	s.processSource(`
	package fixture

	import "gopkg.in/src-d/go-kallax.v1"
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
kallax.NewSchemaField("baz"),
kallax.NewSchemaField("arr"),
kallax.NewSchemaField("arr_aliased"),
kallax.NewSchemaField("urlarr"),
kallax.NewSchemaField("json"),
kallax.NewSchemaField("url"),
kallax.NewSchemaField("url_no_ptr"),
kallax.NewSchemaField("rel_id"),
kallax.NewSchemaField("basic_alias"),
`

func (s *TemplateSuite) TestGenModelColumns() {
	s.processSource(baseTpl)
	m := findModel(s.td.Package, "Foo")
	result := s.td.GenModelColumns(m)
	s.Equal(expectedColumns, result)
}

const jsonBaseTpl = `
	package fixture

	import "gopkg.in/src-d/go-kallax.v1"
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
		Inverse Rel ` + "`fk:\",inverse\"`" + `
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
InverseFK kallax.SchemaField
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
InverseFK:kallax.NewSchemaField("rel_id"),
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

	import "gopkg.in/src-d/go-kallax.v1"
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

	import "gopkg.in/src-d/go-kallax.v1"
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

const expectedTimeTruncations = `record.CreatedAt = record.CreatedAt.Truncate(time.Microsecond)
record.UpdatedAt = record.UpdatedAt.Truncate(time.Microsecond)
record.T = record.T.Truncate(time.Microsecond)
if record.TPtr != nil {
record.TPtr = func(t time.Time) *time.Time { return &t }(record.TPtr.Truncate(time.Microsecond))
}
`

func (s *TemplateSuite) TestGenTimeTruncations() {
	s.processSource(`
	package fixture

	import "gopkg.in/src-d/go-kallax.v1"
	import "time"

	type Foo struct {
		kallax.Model
		ID int64 ` + "`pk:\"autoincr\"`" + `
		kallax.Timestamps
		T time.Time
		TPtr *time.Time
		Foo string
	}
	`)

	m := findModel(s.td.Package, "Foo")
	s.Equal(expectedTimeTruncations, s.td.GenTimeTruncations(m))
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

const (
	equalizable       string = "equalizable"
	sortable                 = "sortable"
	collection               = "collection"
	multiType                = "multy type"
	none                     = "none"
	testSkippedErrMsg        = "Test skipped because testing fixture was not successfully processed"
)

func (s *ProcessorSuite) TestFindableTypeName() {
	fixtureSrc := `
		package foo

		import "time"
		import "net/url"
		import "gopkg.in/src-d/go-kallax.v1"
		import "gopkg.in/src-d/go-kallax.v1/tests/fixtures"

		type mainFixture struct {
			kallax.Model
			ID						kallax.ULID			` + "`findable:\"kallax.ULID\" pk:\"\"`" + `

			StringProp 				string 				` + "`findable:\"string\"`" + `
			SliceStringProp 		[]string 			` + "`findable:\"string\"`" + `
			ArrStringProp 			[2]string 			` + "`findable:\"string\"`" + `
			AliasArrStringProp 		AliasArrString 		` + "`findable:\"string\"`" + `
			AliasStringProp			AliasString 		` + "`findable:\"AliasString\"`" + `
			ArrAliasStringProp		[]AliasString 		` + "`findable:\"AliasString\"`" + `
			AliasArrAliasStringProp AliasArrAliasString	` + "`findable:\"AliasString\"`" + `

			TimeProp				time.Time			` + "`findable:\"time.Time\"`" + `
			ArrTimeProp				[]time.Time 		` + "`findable:\"time.Time\"`" + `
			AliasArrTimeProp		AliasArrTime		` + "`findable:\"time.Time\"`" + `

			IDProp					kallax.ULID			` + "`findable:\"kallax.ULID\"`" + `
			ArrIDProp				[]kallax.ULID			` + "`findable:\"kallax.ULID\"`" + `
			AliasArrIDProp			AliasArrID			` + "`findable:\"kallax.ULID\"`" + `

			IfaceProp 				ScannerValuer 		` + "`findable:\"ScannerValuer\"`" + `
			ArrIfaceProp 			[]ScannerValuer 	` + "`findable:\"ScannerValuer\"`" + `
			AliasArrIfaceProp 		AliasArrIface 		` + "`findable:\"ScannerValuer\"`" + `
			AliasIfaceProp 			AliasIface 			` + "`findable:\"AliasIface\"`" + `
			ArrAliasIfaceProp 		[]AliasIface 		` + "`findable:\"AliasIface\"`" + `
			AliasArrAliasIfaceProp	AliasArrAliasIface 	` + "`findable:\"AliasIface\"`" + `

			UrlProp					url.URL				` + "`findable:\"url.URL\"`" + `
			ArrUrlProp				[]url.URL			` + "`findable:\"url.URL\"`" + `

			ExtAliasIfacePtrProp	*ScannerValuer		` + "`findable:\"ScannerValuer\"`" + `
			ExternalAliasProp		fixtures.AliasInt	` + "`findable:\"fixtures.AliasInt\"`" + `
			ArrExternalAliasProp	[]fixtures.AliasInt	` + "`findable:\"fixtures.AliasInt\"`" + `

			ArrArrStringProp		[][]string
			ArrAliasArrString 		[]AliasArrString
			WhateverProp			Whatever
			ArrWhateverProp 		[]Whatever
		}

		type AliasString string
		type AliasArrString []string
		type AliasArrAliasString []AliasString

		type AliasTime time.Time
		type AliasArrTime []time.Time
		type AliasArrAliasTime []AliasTime

		type AliasArrID []kallax.ULID

		type AliasIface ScannerValuer
		type AliasArrIface []ScannerValuer
		type AliasArrAliasIface []AliasIface

		type ScannerValuer struct {
			fixtures.ScannerValuer
		}

		type Whatever struct {
			name string
		}
	`

	_, model := s.testedModel(fixtureSrc, "mainFixture")
	if model == nil {
		s.Fail(testSkippedErrMsg)
		return
	}
	for _, field := range model.Fields {
		s.assertFindableTypeName(field)
	}
}

func (s *ProcessorSuite) assertFindableTypeName(f *Field) {
	if f.Name == "Model" {
		return
	}

	findableTypeName, ok := findableTypeName(f.Node.Type(), f.Node.Pkg())
	if expected := f.Tag.Get("findable"); expected != "" {
		relativeFindableTypeName := getRelativeTypeName(findableTypeName, "foo")
		s.True(ok, fmt.Sprintf("Could not be found the findable type name of '%s' %s", f.Name, f.Node.Type()))
		s.Equal(expected, relativeFindableTypeName, fmt.Sprintf("Wrong findable type of '%s' %s - %s", f.Name, f.Node.Type(), findableTypeName))
	} else {
		s.False(ok, fmt.Sprintf("Not findable property: '%s' %s", f.Name, f.Node.Type()))
	}
}

func (s *ProcessorSuite) TestLookupValid() {
	fixtureSrc := `
		package foo

		import "time"
		import "net/url"
		import "gopkg.in/src-d/go-kallax.v1"
		import "gopkg.in/src-d/go-kallax.v1/tests/fixtures"

		type mainFixture struct {
			kallax.Model
			ID                   kallax.ULID		` + "`valid:\"kallax.ULID\" type:\"" + equalizable + "\" pk:\"\"`" + `

			StringProp           string				` + "`valid:\"string\" type:\"" + equalizable + "\"`" + `
			ArrStringProp        []string			` + "`deep:\"[]string\" type:\"" + collection + "\"`" + `
			IntProp              int				` + "`valid:\"int\" type:\"" + sortable + "\"`" + `
			ArrIntProp           []int				` + "`deep:\"[]int\" type:\"" + collection + "\"`" + `
			Int64Prop            int64				` + "`valid:\"int64\" type:\"" + sortable + "\"`" + `
			SliceInt64Prop       []int64			` + "`deep:\"[]int64\" type:\"" + collection + "\"`" + `
			ArrInt64Prop         [2]int64			` + "`deep:\"[2]int64\" type:\"" + collection + "\"`" + `
			Float32Prop          float32			` + "`valid:\"float32\" type:\"" + sortable + "\"`" + `
			ArrFloat32Prop       []float32			` + "`deep:\"[]float32\" type:\"" + collection + "\"`" + `
			Uint8Prop            uint8				` + "`valid:\"uint8\" type:\"" + sortable + "\"`" + `
			ArrUint8Prop         []uint8			` + "`deep:\"[]uint8\" type:\"" + collection + "\"`" + `
			BoolProp             bool				` + "`valid:\"bool\" type:\"" + equalizable + "\"`" + `
			ArrBoolProp          []bool				` + "`deep:\"[]bool\" type:\"" + collection + "\"`" + `
			ByteProp             byte				` + "`valid:\"byte\" type:\"" + sortable + "\"`" + `
			ArrByteProp          []byte				` + "`deep:\"[]byte\" type:\"" + collection + "\"`" + `
			IDProp               kallax.ULID		` + "`valid:\"kallax.ULID\" type:\"" + equalizable + "\"`" + `
			ArrIDProp            []kallax.ULID		` + "`deep:\"[]kallax.ULID\" type:\"" + collection + "\"`" + `
			UrlProp              url.URL			` + "`valid:\"url.URL\" type:\"" + equalizable + "\"`" + `
			ArrUrlProp           []url.URL			` + "`deep:\"[]url.URL\" type:\"" + collection + "\"`" + `
			TimeProp             time.Time			` + "`valid:\"time.Time\" type:\"" + sortable + "\"`" + `
			ArrTimeProp          []time.Time		` + "`deep:\"[]time.Time\" type:\"" + collection + "\"`" + `
			AliasStringProp      AliasString		` + "`valid:\"string\" type:\"" + equalizable + "\"`" + `
			AliasAliasStringProp AliasAliasString	` + "`valid:\"string\" type:\"" + equalizable + "\"`" + `
			AliasArrStringProp   AliasArrString		` + "`deep:\"[]string\" type:\"" + collection + "\"`" + `
			IfaceProp            EmbScannerValuer	` + "`valid:\"EmbScannerValuer\" type:\"" + equalizable + "\"`" + `
			AliasEmbIfaceProp    AliasEmbIface		` + "`valid:\"AliasEmbIface\" type:\"" + equalizable + "\"`" + `
			AliasIfaceProp       AliasIface			` + "`deep:\"struct{}\" type:\"" + none + "\"`" + `
			ArrWhateverProp      []Whatever			` + "`deep:\"[]Whatever\" type:\"" + collection + "\"`" + `
		}

		type AliasString string
		type AliasAliasString AliasString
		type AliasArrString []string
		type AliasIface fixtures.ScannerValuer
		type AliasEmbIface EmbScannerValuer

		type EmbScannerValuer struct {
			fixtures.ScannerValuer
		}

		type Whatever struct {
			name string
		}
	`

	pkg, model := s.testedModel(fixtureSrc, "mainFixture")
	if model == nil {
		s.Fail(testSkippedErrMsg)
		return
	}
	for _, field := range model.Fields {
		s.assertLookupValid(field, pkg.pkg)
		s.assertTypeOfFindBy(field)
	}
}

func (s *ProcessorSuite) assertLookupValid(f *Field, pkg *types.Package) {
	if f.Name == "Model" {
		return
	}

	expectedValid := f.Tag.Get("valid")
	expectedDeep := f.Tag.Get("deep")
	receivedValid, receivedDeep := lookupValid(f.Node.Pkg(), f.Node.Type())
	switch {
	case expectedValid != "" && receivedValid != nil && receivedDeep == nil:
		validShortName := getRelativeTypeName(shortName(pkg, receivedValid), "foo")
		s.Equal(expectedValid, validShortName, fmt.Sprintf(
			"Wrong valid type of '%s' %s\nreceived: %s",
			f.Name, f.Node.Type(), receivedValid,
		))
	case expectedDeep != "" && receivedValid == nil && receivedDeep != nil:
		deepShortName := getRelativeTypeName(shortName(pkg, receivedDeep), "foo")
		s.Equal(expectedDeep, deepShortName, fmt.Sprintf(
			"Wrong deepest underlying type of '%s' %s\nreceived: %s",
			f.Name, f.Node.Type(), receivedDeep,
		))
	default:
		var recValidName, recDeepestName string
		if receivedValid != nil {
			recValidName = receivedValid.String()
		}
		if receivedDeep != nil {
			recDeepestName = receivedDeep.String()
		}
		summary := fmt.Sprintf("- valid: %s\n- deepest: %s", recValidName, recDeepestName)
		s.Fail(fmt.Sprintf(
			"It must be received only a valid or a deep type for '%s' %s\n%s",
			f.Name, f.Node.Type(), summary,
		))
	}
}

func (s *ProcessorSuite) assertTypeOfFindBy(f *Field) {
	if f.Name == "Model" {
		return
	}
	expectedFindByType := f.Tag.Get("type")
	errorString := fmt.Sprintf("Wrong type for field '%s' %s", f.Name, f.Node.Type())
	s.True(expectedFindByType != "" && findByType(f) != "", errorString)
	s.Equal(expectedFindByType, findByType(f), errorString)
}

func (s *ProcessorSuite) TestShortName() {
	fixtureSrc := `
		package foo

		import "time"
		import "net/url"
		import "gopkg.in/src-d/go-kallax.v1"
		import "gopkg.in/src-d/go-kallax.v1/tests/fixtures"

		type mainFixture struct {
			kallax.Model
			ID                   kallax.ULID		` + "`short:\"kallax.ULID\" pk:\"\"`" + `

			StringProp           string				` + "`short:\"string\"`" + `
			SliceStringProp      []string			` + "`short:\"[]string\"`" + `
			ArrStringProp        [2]string			` + "`short:\"[2]string\"`" + `
			IDProp             	 kallax.ULID		` + "`short:\"kallax.ULID\"`" + `
			UrlProp            	 url.URL			` + "`short:\"url.URL\"`" + `
			TimeProp             time.Time			` + "`short:\"time.Time\"`" + `
			AliasStringProp      AliasString		` + "`short:\"AliasString\"`" + `
			ArrAliasStringProp   []AliasString		` + "`short:\"[]AliasString\"`" + `
			ExternalAliasProp    fixtures.AliasInt	` + "`short:\"fixtures.AliasInt\"`" + `
			ArrExternalAliasProp []fixtures.AliasInt` + "`short:\"[]fixtures.AliasInt\"`" + `
			AliasStringPtrProp   *AliasString		` + "`short:\"AliasString\"`" + `
			ExternalAliasPtrProp *fixtures.AliasInt	` + "`short:\"fixtures.AliasInt\"`" + `
		}

		type AliasString string
	`

	pkg, model := s.testedModel(fixtureSrc, "mainFixture")
	if model == nil {
		s.Fail(testSkippedErrMsg)
		return
	}
	for _, field := range model.Fields {
		s.assertShortName(field, pkg.pkg)
	}
}

func (s *ProcessorSuite) assertShortName(f *Field, pkg *types.Package) {
	if f.Name == "Model" {
		return
	}
	expected := f.Tag.Get("short")
	processedShortName := getRelativeTypeName(shortName(pkg, f.Node.Type()), "foo")
	s.Equal(expected, processedShortName, fmt.Sprintf("Wrong shortName type of '%s' %s", f.Name, f.Node.Type()))
}

func (s *ProcessorSuite) testedModel(fixture, name string) (*Package, *Model) {
	pkg := s.processFixture(fixture)
	s.NotNil(pkg)
	model := findModel(pkg, name)
	s.NotNil(model)
	return pkg, model
}

func getRelativeTypeName(typeName, relativeToPackageName string) string {
	return strings.Replace(typeName, relativeToPackageName+".", "", 1)
}

func findByType(f *Field) string {
	switch {
	case isEqualizable(f) && !isSortable(f) && !isCollection(f):
		return equalizable
	case !isEqualizable(f) && isSortable(f) && !isCollection(f):
		return sortable
	case !isEqualizable(f) && !isSortable(f) && isCollection(f):
		return collection
	case !isEqualizable(f) && !isSortable(f) && !isCollection(f):
		return none
	default:
		return multiType
	}
}
