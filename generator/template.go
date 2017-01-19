package generator

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

type Template struct {
	template *template.Template
}

type TemplateData struct {
	*Package
	Processed  map[interface{}]string
	subschemas map[string]*Field
}

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

func (td *TemplateData) GenColumnAddresses(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsColumnAddresses(&buf, model.Fields)
	return buf.String()
}

func (td *TemplateData) genFieldsColumnAddresses(buf *bytes.Buffer, fields []*Field) {
	for _, f := range fields {
		if f.Kind == Relationship {
			continue
		}

		if f.Inline() {
			td.genFieldsColumnAddresses(buf, f.Fields)
		} else {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ColumnName()))
			buf.WriteString(fmt.Sprintf("return %s\n", f.Address()))
		}
	}
}

func (td *TemplateData) GenColumnValues(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsValues(&buf, model.Fields)
	return buf.String()
}

func (td *TemplateData) genFieldsValues(buf *bytes.Buffer, fields []*Field) {
	for _, f := range fields {
		if f.Kind == Relationship {
			continue
		}

		if f.Inline() {
			td.genFieldsValues(buf, f.Fields)
		} else {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ColumnName()))
			buf.WriteString(fmt.Sprintf("return %s\n", f.Value()))
		}
	}
}

func (td *TemplateData) GenModelColumns(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsColumns(&buf, model.Fields)
	return buf.String()
}

func (td *TemplateData) genFieldsColumns(buf *bytes.Buffer, fields []*Field) {
	for _, f := range fields {
		if f.Kind == Relationship {
			continue
		}

		if f.Inline() {
			td.genFieldsColumns(buf, f.Fields)
		} else {
			buf.WriteString(fmt.Sprintf("kallax.NewSchemaField(\"%s\"),\n", f.ColumnName()))
		}
	}
}

func (td *TemplateData) GenModelSchema(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsSchema(&buf, model.Name, model.Fields)
	return buf.String()
}

func (td *TemplateData) genFieldsSchema(buf *bytes.Buffer, parent string, fields []*Field) {
	for _, f := range fields {
		if f.Kind == Relationship {
			return
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

func (td *TemplateData) GenSubSchemas() string {
	var buf bytes.Buffer
	for parent, field := range td.subschemas {
		buf.WriteString("type schema" + parent + " struct {\n")
		buf.WriteString("*kallax.BaseSchemaField\n")
		td.genFieldsSchema(&buf, parent, field.Fields)
		buf.WriteString("}\n\n")
	}
	return buf.String()
}

func (td *TemplateData) GenSchemaInit(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsInit(&buf, model.Name, model.Fields, true)
	return buf.String()
}

func (td *TemplateData) genFieldsInit(buf *bytes.Buffer, parent string, fields []*Field, root bool) {
	for _, f := range fields {
		if f.Kind == Relationship {
			return
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
				td.genFieldsInit(buf, parent+f.Name, f.Fields, false)
				buf.WriteString("},")
			} else {
				buf.WriteString(fmt.Sprintf(`kallax.NewSchemaField("%s"),`, schemaName))
			}

			buf.WriteRune('\n')
		}
	}
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

var Base *Template = &Template{template: base}
