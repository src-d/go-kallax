package generator

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

type Template struct {
	template *template.Template
}

type TemplateData struct {
	*Package
	Fields    []*TemplateField
	Processed map[interface{}]string
}

type TemplateField struct {
	Name   string
	Path   string
	Fields interface{}
}

func (tf *TemplateField) ValidFields() []*Field {
	return tf.Fields.([]*Field)
}

func (t *Template) Execute(wr io.Writer, data *Package) error {
	var buf bytes.Buffer

	td := &TemplateData{data, []*TemplateField{}, map[interface{}]string{}}
	err := t.template.Execute(&buf, td)
	if err != nil {
		return err
	}

	return prettyfy(buf.Bytes(), wr)
}

func (td *TemplateData) GenType(vi interface{}, path string) string {
	v := reflect.ValueOf(vi)
	sv := v
	if v.Kind() == reflect.Ptr {
		sv = v.Elem()
	}
	if sv.FieldByName("Type").Interface().(string) == "struct" {
		if v.MethodByName("ValidFields").IsValid() {
			return td.LinkStruct(path, vi)
		}
		return ""
	} else {
		k := "Field"
		if v.MethodByName("ContainsMap").Call(nil)[0].Interface().(bool) {
			k = "Map"
		}
		return fmt.Sprintf("%v storable.%v", sv.FieldByName("Name"), k)
	}
}

func (td *TemplateData) LinkStruct(path string, vi interface{}) string {
	v := reflect.ValueOf(vi)
	name := v.Elem().FieldByName("Name").Interface().(string)
	schemaName := "schema" + path + name

	if proc, ok := td.Processed[vi]; ok {
		schemaName = proc
		return name + " *" + schemaName
	}
	td.Processed[vi] = schemaName

	td.Fields = append(td.Fields, &TemplateField{
		Name:   schemaName,
		Path:   path + name,
		Fields: v.MethodByName("ValidFields").Call(nil)[0].Interface(),
	})

	return name + " *" + schemaName
}

func (td *TemplateData) GenVar(vi interface{}, done map[interface{}]bool) string {
	if done == nil {
		done = map[interface{}]bool{}
	}

	v := reflect.ValueOf(vi)
	sv := v
	if v.Kind() == reflect.Ptr {
		sv = v.Elem()
	}

	if done[vi] {
		return sv.FieldByName("Name").Interface().(string) + ": nil,"
	}

	if sv.FieldByName("Type").Interface().(string) == "struct" {
		if v.MethodByName("ValidFields").IsValid() {
			return td.StructValue(vi, done)
		}
		return ""
	} else {
		k := "NewField"
		if v.MethodByName("ContainsMap").Call(nil)[0].Interface().(bool) {
			k = "NewMap"
		}

		return fmt.Sprintf(
			`%v: storable.%v("%v", "%v"),`,
			sv.FieldByName("Name"),
			k,
			v.MethodByName("GetPath").Call(nil)[0],
			v.MethodByName("FindableType").Call(nil)[0],
		)
	}
}

func (td *TemplateData) StructValue(vi interface{}, done map[interface{}]bool) string {
	v := reflect.ValueOf(vi)
	name := v.Elem().FieldByName("Name").Interface().(string)

	ifc := v.Interface()
	if done[ifc] {
		return name + ": nil,"
	}
	done[ifc] = true

	ret := name + ": &" + td.Processed[vi] + "{"
	for _, v := range v.MethodByName("ValidFields").Call(nil)[0].Interface().([]*Field) {
		ret += "\n" + td.GenVar(v, done)
	}
	ret += "\n},"

	return ret
}

func prettyfy(input []byte, wr io.Writer) error {
	output, err := imports.Process("storable.go", input, nil)
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
	filename = filepath.Join(build.Default.GOPATH, "src/gopkg.in/src-d/storable.v1/generator", filename)
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
