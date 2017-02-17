package generator

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

// Template renders the kallax templates using given packages.
type Template struct {
	template *template.Template
}

// TemplateData is the structure passed to fill the templates.
type TemplateData struct {
	*Package
	// Processed is a map to keep track of processed nodes.
	Processed  map[interface{}]string
	subschemas map[string]*Field
}

// Execute writes the processed template to the given writer.
func (t *Template) Execute(wr io.Writer, data *Package) error {
	var buf bytes.Buffer

	td := &TemplateData{
		data,
		map[interface{}]string{},
		map[string]*Field{},
	}
	err := t.template.Execute(&buf, td)
	if err != nil {
		return err
	}

	return prettyfy(buf.Bytes(), wr)
}

// GenColumnAddresses generates the body of the switch that returns the column
// address given a column name for the given model.
func (td *TemplateData) GenColumnAddresses(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsColumnAddresses(&buf, model.Fields)
	return buf.String()
}

const initNilPtrTpl = `if r.%s == nil {
r.%s = new(%s)
}
`

func (td *TemplateData) genFieldsColumnAddresses(buf *bytes.Buffer, fields []*Field) {
	for _, f := range fields {
		if f.Inline() {
			td.genFieldsColumnAddresses(buf, f.Fields)
		} else if f.Kind == Relationship && f.IsInverse() {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ForeignKey()))
			buf.WriteString(fmt.Sprintf("return kallax.VirtualColumn(\"%s\", r, new(%s)), nil\n", f.ForeignKey(), td.foreignKeyType(f)))
		} else if f.Kind != Relationship {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ColumnName()))
			if f.IsPrimaryKey() {
				buf.WriteString(fmt.Sprintf("return (*%s)(%s), nil\n", td.IdentifierType(f), f.fieldVarAddress()))
			} else {
				// can't scan a json if is nil
				if f.IsJSON && f.IsPtr {
					buf.WriteString(fmt.Sprintf(initNilPtrTpl, f.Name, f.Name, td.GenTypeName(f)))
				}
				buf.WriteString(fmt.Sprintf("return %s, nil\n", f.Address()))
			}
		}
	}
}

func (td *TemplateData) foreignKeyType(f *Field) string {
	model := td.Package.FindModel(f.TypeSchemaName())
	return identifierType(model.ID)
}

func (td *TemplateData) IdentifierType(f *Field) string {
	return identifierType(f)
}

// GenColumnValues generates the body of the switch that returns the column
// address given a column name for the given model.
func (td *TemplateData) GenColumnValues(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsValues(&buf, model.Fields)
	return buf.String()
}

func (td *TemplateData) genFieldsValues(buf *bytes.Buffer, fields []*Field) {
	for _, f := range fields {
		if f.Inline() {
			td.genFieldsValues(buf, f.Fields)
		} else if f.Kind == Relationship && f.IsInverse() {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ForeignKey()))
			buf.WriteString(fmt.Sprintf("return r.Model.VirtualColumn(col), nil\n"))
		} else if f.Kind != Relationship {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ColumnName()))
			buf.WriteString(fmt.Sprintf("return %s\n", f.Value()))
		}
	}
}

// GenModelColumns generates the creation of the list of columns in the given
// model.
func (td *TemplateData) GenModelColumns(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsColumns(&buf, model.Fields)
	return buf.String()
}

func (td *TemplateData) genFieldsColumns(buf *bytes.Buffer, fields []*Field) {
	for _, f := range fields {
		if f.Inline() {
			td.genFieldsColumns(buf, f.Fields)
		} else if f.Kind == Relationship && f.IsInverse() {
			buf.WriteString(fmt.Sprintf("kallax.NewSchemaField(\"%s\"),\n", f.ForeignKey()))
		} else if f.Kind != Relationship {
			buf.WriteString(fmt.Sprintf("kallax.NewSchemaField(\"%s\"),\n", f.ColumnName()))
		}
	}
}

// GenModelSchema generates generates the fields of the struct definition
// in the given model.
func (td *TemplateData) GenModelSchema(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsSchema(&buf, model.Name, model.Fields)
	return buf.String()
}

func (td *TemplateData) genFieldsSchema(buf *bytes.Buffer, parent string, fields []*Field) {
	for _, f := range fields {
		if f.Kind == Relationship {
			continue
		}

		if f.Inline() {
			td.genFieldsSchema(buf, parent, f.Fields)
		} else {
			buf.WriteString(f.Name + " ")

			if f.IsJSON && len(f.Fields) > 0 {
				buf.WriteString("*schema" + parent + f.Name)
				td.findJSONSchemas(parent, f)
			} else {
				buf.WriteString("kallax.SchemaField")
			}

			buf.WriteRune('\n')
		}
	}
}

func (td *TemplateData) findJSONSchemas(parent string, f *Field) {
	n := parent + f.Name
	if _, ok := td.subschemas[n]; ok {
		return
	}

	td.subschemas[n] = f

	for _, f := range f.Fields {
		if f.IsJSON && len(f.Fields) > 0 {
			td.findJSONSchemas(n, f)
		}
	}
}

// GenTypeName generates the name of the type in the field.
func (td *TemplateData) GenTypeName(f *Field) string {
	return removeTypePrefix(typeString(f.Node.Type(), td.pkg))
}

// IsPtrSlice returns whether the field is a slice of pointers or not.
func (td *TemplateData) IsPtrSlice(f *Field) bool {
	return strings.HasPrefix(typeString(f.Node.Type(), td.pkg), "[]*")
}

func removeTypePrefix(typ string) string {
	return strings.TrimLeft(typ, "[]*")
}

// GenSubSchemas generates the struct definition of all the subschemas in all
// models.
// A subschema is the JSON schema of a field that will be stored as JSON.
func (td *TemplateData) GenSubSchemas() string {
	var buf bytes.Buffer

	var names = make([]string, 0, len(td.subschemas))
	for n := range td.subschemas {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		field := td.subschemas[name]
		buf.WriteString("type schema" + name + " struct {\n")
		if isSliceOrArray(field) {
			if isRootField(field) {
				buf.WriteString("*kallax.BaseSchemaField\n")
			} else {
				buf.WriteString("*kallax.JSONSchemaArray\n")
			}
		} else {
			buf.WriteString("*kallax.BaseSchemaField\n")
		}
		td.genFieldsSchema(&buf, name, field.Fields)
		buf.WriteString("}\n\n")

		if isSliceOrArray(field) {
			td.genArraySchemaAtFunc(&buf, name, field)
		}
	}
	return buf.String()
}

// genArraySchemaAtFunc generates the `At` func for an array field schema.
func (td *TemplateData) genArraySchemaAtFunc(buf *bytes.Buffer, parent string, f *Field) {
	buf.WriteString(fmt.Sprintf("func (s *schema%s) At(n int) *schema%s {\n", parent, parent))
	buf.WriteString(fmt.Sprintf("return &schema%s{\n", parent))

	if isRootField(f) {
		buf.WriteString(fmt.Sprintf("BaseSchemaField: kallax.NewSchemaField(%s).(*kallax.BaseSchemaField),\n", td.genSchemaPath(f)))
	} else {
		buf.WriteString(fmt.Sprintf("JSONSchemaArray: kallax.NewJSONSchemaArray(%s),\n", td.genSchemaPath(f)))
	}

	td.genSubschemaFieldsInit(buf, parent, f.Fields, "fmt.Sprint(n)")
	buf.WriteString("}\n}\n\n")
}

// genSchemaPath generates the path needed to access the given field in JSON.
// If prependLast is given, it will add the items before the last element of the path.
// It also returns a boolean reporting whether the path has a single level of
// depth (that is, we're talking about the column itself).
func (td *TemplateData) genSchemaPath(f *Field, prependLast ...string) string {
	var result string
	for f.Parent != nil {
		if !f.Inline() {
			if result == "" {
				result = fmt.Sprintf("%q", f.JSONName())
				if len(prependLast) > 0 {
					result = fmt.Sprintf("%s, %s", strings.Join(prependLast, ", "), result)
				}
			} else {
				result = fmt.Sprintf("%q, %s", f.JSONName(), result)
			}
		}

		f = f.Parent
	}

	result = fmt.Sprintf("%q, %s", f.ColumnName(), result)
	return strings.TrimRight(strings.TrimSpace(result), ",")
}

// genSubschemaFieldsInit generates the initialization of the subschema fields.
// If prependLast is given, all field paths will have prependLast before the
// last element of its path.
func (td *TemplateData) genSubschemaFieldsInit(buf *bytes.Buffer, parent string, fields []*Field, prependLast string) {
	for _, f := range fields {
		if f.Inline() {
			td.genSubschemaFieldsInit(buf, parent, f.Fields, "")
		} else {
			buf.WriteString(fmt.Sprintf("%s:", f.Name))

			var path string
			if prependLast != "" {
				path = td.genSchemaPath(f, prependLast)
			} else {
				path = td.genSchemaPath(f)
			}

			if f.IsJSON && len(f.Fields) > 0 {
				td.genSubschemaInit(buf, parent, f)
			} else if isSliceOrArray(f) {
				buf.WriteString(fmt.Sprintf("kallax.NewJSONSchemaArray(%s)", path))
			} else {
				buf.WriteString(fmt.Sprintf(
					"kallax.NewJSONSchemaKey(%s, %s)",
					td.genJSONType(f),
					path,
				))
			}
			buf.WriteString(",\n")
		}
	}
}

// genSubschemaInit generates the initialization for a subschema.
func (td *TemplateData) genSubschemaInit(buf *bytes.Buffer, parent string, f *Field) {
	buf.WriteString(fmt.Sprintf("&schema%s%s{\n", parent, f.Name))
	if isSliceOrArray(f) {
		if isRootField(f) {
			buf.WriteString(fmt.Sprintf("BaseSchemaField: kallax.NewSchemaField(%s).(*kallax.BaseSchemaField),\n", td.genSchemaPath(f)))
		} else {
			buf.WriteString(fmt.Sprintf("JSONSchemaArray: kallax.NewJSONSchemaArray(%s),\n", td.genSchemaPath(f)))
		}
	} else {
		buf.WriteString(fmt.Sprintf(
			"JSONSchemaKey: kallax.NewJSONSchemaKey(%s, %s),\n",
			td.genJSONType(f),
			td.genSchemaPath(f),
		))
	}
	td.genSubschemaFieldsInit(buf, parent+f.Name, f.Fields, "")
	buf.WriteString("}")
}

// GenSchemaInit generates the code to initialize all fields in the schema
// of a model.
func (td *TemplateData) GenSchemaInit(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsInit(&buf, model.Name, model.Fields, true)
	return buf.String()
}

func (td *TemplateData) genFieldsInit(buf *bytes.Buffer, parent string, fields []*Field, root bool) {
	for _, f := range fields {
		if f.Kind == Relationship {
			continue
		}

		if f.Inline() {
			td.genFieldsInit(buf, parent, f.Fields, true)
		} else {
			buf.WriteString(f.Name + ":")
			var schemaName = f.Name
			if root {
				schemaName = f.ColumnName()
			}

			if f.IsJSON && len(f.Fields) > 0 {
				buf.WriteString(fmt.Sprintf("&schema%s%s{\n", parent, f.Name))
				buf.WriteString(fmt.Sprintf(`BaseSchemaField: kallax.NewSchemaField("%s").(*kallax.BaseSchemaField),`+"\n", schemaName))
				td.genSubschemaFieldsInit(buf, parent+f.Name, f.Fields, "")
				buf.WriteString("},")
			} else {
				buf.WriteString(fmt.Sprintf(`kallax.NewSchemaField("%s"),`, schemaName))
			}

			buf.WriteRune('\n')
		}
	}
}

func (td *TemplateData) genJSONType(f *Field) string {
	switch f.Type {
	case "string":
		return "kallax.JSONText"
	case "int8", "uint8", "byte", "int16", "uint16", "int32", "uint32", "int", "uint", "int64", "uint64":
		return "kallax.JSONInt"
	case "float64", "float32":
		return "kallax.JSONFloat"
	case "bool":
		return "kallax.JSONBool"
	default:
		return "kallax.JSONAny"
	}
}

// isRootField reports whether the field is at the top level of the model.
// It takes into account if the parent is inlined or not.
func isRootField(f *Field) bool {
	return f.Parent == nil || (f.Parent.Inline() && isRootField(f.Parent))
}

func isSliceOrArray(f *Field) bool {
	return strings.HasPrefix(f.Type, "[")
}

func prettyfy(input []byte, wr io.Writer) error {
	output, err := imports.Process("kallax.go", input, nil)
	if err != nil {
		printDocumentWithNumbers(string(input))
		return err
	}

	_, err = wr.Write(output)
	return err
}

func printDocumentWithNumbers(code string) {
	for i, line := range strings.Split(code, "\n") {
		fmt.Printf("%.3d %s\n", i+1, line)
	}
}

func loadTemplateText(filename string) string {
	filename = filepath.Join(build.Default.GOPATH, "src/github.com/src-d/go-kallax/generator", filename)
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(f); err != nil {
		panic(err)
	}

	return strings.Replace(buf.String(), "\\\n", " ", -1)
}

func makeTemplate(name string, filename string) *template.Template {
	text := loadTemplateText(filename)
	return template.Must(template.New(name).Parse(text))
}

func addTemplate(base *template.Template, name string, filename string) *template.Template {
	text := loadTemplateText(filename)
	return template.Must(base.New(name).Parse(text))
}

var base *template.Template = makeTemplate("base", "templates/base.tgo")
var schema *template.Template = addTemplate(base, "schema", "templates/schema.tgo")
var model *template.Template = addTemplate(base, "model", "templates/model.tgo")
var query *template.Template = addTemplate(model, "query", "templates/query.tgo")
var resultset *template.Template = addTemplate(model, "resultset", "templates/resultset.tgo")

// Base is the default Template instance with all templates preloaded.
var Base *Template = &Template{template: base}
