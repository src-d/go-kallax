package generator

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	parseutil "gopkg.in/src-d/go-parse-utils.v1"

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

func (td *TemplateData) GenTimeTruncations(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsTimeTruncations(&buf, model.Fields)
	return buf.String()
}

const truncateTimePtrTpl = `if record.%s != nil {
record.%s = func(t time.Time) *time.Time { return &t }(record.%s.Truncate(time.Microsecond))
}
`

func (td *TemplateData) genFieldsTimeTruncations(buf *bytes.Buffer, fields []*Field) {
	for _, f := range fields {
		if f.Inline() {
			td.genFieldsTimeTruncations(buf, f.Fields)
			continue
		}

		typ := removeTypePrefix(typeName(f.Node.Type()))
		if typ == "time.Time" {
			if !f.IsPtr {
				buf.WriteString(fmt.Sprintf("record.%s = record.%s.Truncate(time.Microsecond)\n", f.Name, f.Name))
			} else {
				buf.WriteString(fmt.Sprintf(truncateTimePtrTpl, f.Name, f.Name, f.Name))
			}
		}
	}
}

// GenColumnAddresses generates the body of the switch that returns the column
// address given a column name for the given model.
func (td *TemplateData) GenColumnAddresses(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsColumnAddresses(&buf, model.Fields)
	for _, fk := range model.ImplicitFKs {
		buf.WriteString(fmt.Sprintf("case \"%s\":\n", fk.Name))
		buf.WriteString(fmt.Sprintf("return types.Nullable(kallax.VirtualColumn(\"%s\", r, new(%s))), nil\n", fk.Name, fk.Type))
	}
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
		} else if isOneToOneRelationship(f) && f.IsInverse() {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ForeignKey()))
			buf.WriteString(fmt.Sprintf("return types.Nullable(kallax.VirtualColumn(\"%s\", r, new(%s))), nil\n", f.ForeignKey(), td.foreignKeyType(f)))
		} else if f.Kind != Relationship {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ColumnName()))
			if f.IsPrimaryKey() {
				buf.WriteString(fmt.Sprintf("return (*%s)(%s), nil\n", td.IdentifierType(f), f.fieldVarAddress()))
			} else {
				// can't scan a json if is nil
				if (f.IsJSON || f.Kind == Interface) && f.IsPtr {
					buf.WriteString(fmt.Sprintf(initNilPtrTpl, f.Name, f.Name, td.GenTypeName(f)))
				}

				if f.Kind == Basic && f.IsAlias {
					buf.WriteString(fmt.Sprintf("return (*%s)(%s), nil\n", f.Type, f.Address()))
				} else {
					buf.WriteString(fmt.Sprintf("return %s, nil\n", f.Address()))
				}
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
	for _, fk := range model.ImplicitFKs {
		buf.WriteString(fmt.Sprintf("case \"%s\":\n", fk.Name))
		buf.WriteString(fmt.Sprintf("return r.Model.VirtualColumn(col), nil\n"))
	}
	return buf.String()
}

const nilPtrReturnsUntypedNilTpl = `if %s == (*%s)(nil) {
	return nil, nil
}
`

func (td *TemplateData) genFieldsValues(buf *bytes.Buffer, fields []*Field) {
	for _, f := range fields {
		if f.Inline() {
			td.genFieldsValues(buf, f.Fields)
		} else if isOneToOneRelationship(f) && f.IsInverse() {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ForeignKey()))
			buf.WriteString(fmt.Sprintf("return r.Model.VirtualColumn(col), nil\n"))
		} else if f.Kind != Relationship {
			buf.WriteString(fmt.Sprintf("case \"%s\":\n", f.ColumnName()))
			if f.IsPtr {
				buf.WriteString(fmt.Sprintf(nilPtrReturnsUntypedNilTpl, f.fieldVarName(), td.GenTypeName(f)))
			}
			buf.WriteString(fmt.Sprintf("return %s\n", f.Value()))
		}
	}
}

// GenModelColumns generates the creation of the list of columns in the given
// model.
func (td *TemplateData) GenModelColumns(model *Model) string {
	var buf bytes.Buffer
	td.genFieldsColumns(&buf, model.Fields)
	for _, fk := range model.ImplicitFKs {
		buf.WriteString(fmt.Sprintf("kallax.NewSchemaField(\"%s\"),\n", fk.Name))
	}
	return buf.String()
}

func (td *TemplateData) genFieldsColumns(buf *bytes.Buffer, fields []*Field) {
	for _, f := range fields {
		if f.Inline() {
			td.genFieldsColumns(buf, f.Fields)
		} else if isOneToOneRelationship(f) && f.IsInverse() {
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
		if f.Kind == Relationship && !f.IsInverse() {
			continue
		}

		if f.Inline() {
			td.genFieldsSchema(buf, parent, f.Fields)
		} else if isOneToOneRelationship(f) && f.IsInverse() {
			buf.WriteString(fmt.Sprintf("%sFK kallax.SchemaField\n", f.Name))
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
	if name, ok := findNamed(f.Node.Type(), td.pkg); ok {
		return name
	}

	return removeTypePrefix(typeString(f.Node.Type(), td.pkg))
}

func findNamed(t types.Type, pkg *types.Package) (string, bool) {
	switch t := t.(type) {
	case *types.Pointer:
		return findNamed(t.Elem(), pkg)
	case *types.Named:
		if t.Obj().Pkg().Path() == pkg.Path() {
			return t.Obj().Name(), true
		}

		return fmt.Sprintf("%s.%s", t.Obj().Pkg().Name(), t.Obj().Name()), true
	default:
		return "", false
	}
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
		if f.Kind == Relationship && !f.IsInverse() {
			continue
		}

		if f.Inline() {
			td.genFieldsInit(buf, parent, f.Fields, true)
		} else if isOneToOneRelationship(f) && f.IsInverse() {
			buf.WriteString(fmt.Sprintf("%sFK:kallax.NewSchemaField(\"%s\"),\n", f.Name, f.ForeignKey()))
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

	out := strings.Replace(string(output), "{\n\n", "{\n", -1)
	out = strings.Replace(out, "\n\n}", "\n}", -1)

	_, err = wr.Write([]byte(out))
	return err
}

func printDocumentWithNumbers(code string) {
	for i, line := range strings.Split(code, "\n") {
		fmt.Printf("%.3d %s\n", i+1, line)
	}
}

const pkgPath = "gopkg.in/src-d/go-kallax.v1/generator"

var pkgAbsPath = func() string {
	path, err := parseutil.DefaultGoPath.Abs(pkgPath)
	if err != nil {
		panic(err)
	}
	return path
}()

func loadTemplateText(filename string) string {
	filename = filepath.Join(pkgAbsPath, filename)
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

const (
	// tplFindByCollection is the template of the FindBy autogenerated for
	// properties that are collection.
	// The passed values to the FindBy will be used in an kallax.ArrayContains
	tplFindByCollection = `
		// FindBy%[1]s adds a new filter to the query that will require that
		// the %[1]s property contains all the passed values; if no passed values, 
		// it will do nothing.
		func (q *%[2]s) FindBy%[1]s(v ...%[3]s) *%[2]s {
		    if len(v) == 0 {return q}
		    values := make([]interface{}, len(v))
		    for i, val := range v {values[i] = val}
		    return q.Where(kallax.ArrayContains(Schema.%[4]s.%[1]s, values...))
		}`
	// tplFindByEquality is the template of the FindBy autogenerated for
	// properties that will be searched with an kallax.Eq condition.
	tplFindByEquality = `
		// FindBy%[1]s adds a new filter to the query that will require that
		// the %[1]s property is equal to the passed value.
		func (q *%[2]s) FindBy%[1]s(v %[3]s) *%[2]s {
			return q.Where(kallax.Eq(Schema.%[4]s.%[1]s, v))
		}`
	// tplFindByCondition is the template of the FindBy autogenerated for
	// properties that can be compared regarding to a kallax.ScalarCond condition.
	tplFindByCondition = `
		// FindBy%[1]s adds a new filter to the query that will require that
		// the %[1]s property is equal to the passed value.
		func (q *%[2]s) FindBy%[1]s(cond kallax.ScalarCond, v %[3]s) *%[2]s {
			return q.Where(cond(Schema.%[4]s.%[1]s, v))
		}`
	// tplFindByID is the template of the FindBy autogenerated for the primary key.
	// The passed values to the FindBy will be used in an kallax.In condition.
	tplFindByID = `
		// FindBy%[1]s adds a new filter to the query that will require that
		// the %[1]s property is equal to one of the passed values; if no passed values, 
		// it will do nothing.
		func (q *%[2]s) FindBy%[1]s(v ...%[3]s) *%[2]s {
			if len(v) == 0 {return q}
			values := make([]interface{}, len(v))
			for i, val := range v {values[i] = val}
			return q.Where(kallax.In(Schema.%[4]s.%[1]s, values...))
		}`
	// tplFindByFK is the template of the FindBy autogenerated for the primary key.
	// The passed values to the FindBy will be used in an kallax.In condition.
	tplFindByFK = `
		// FindBy%[1]s adds a new filter to the query that will require that
		// the foreign key of %[1]s is equal to the passed value.
		func (q *%[2]s) FindBy%[1]s(v %[3]s) *%[2]s {
			return q.Where(kallax.Eq(Schema.%[4]s.%[1]sFK, v))
		}`
)

// GenFindBy generates FindByPropertyName for all model properties that are
// valid types or collection of valid types.
func (td *TemplateData) GenFindBy(model *Model) string {
	var buf bytes.Buffer
	td.genFindBy(&buf, model, model.Fields)
	return buf.String()
}

func (td *TemplateData) genFindBy(buf *bytes.Buffer, parent *Model, fields []*Field) {
	for _, f := range fields {
		switch {
		case f.Inline():
			td.genFindBy(buf, parent, f.Fields)
		case f.IsPrimaryKey():
			writeFindByTpl(buf, parent, f.Name, f, tplFindByID)
		case isOneToOneRelationship(f) && f.IsInverse():
			model := td.FindModel(f.TypeSchemaName())
			writeFindByTpl(buf, parent, f.Name, model.ID, tplFindByFK)
		case isEqualizable(f):
			writeFindByTpl(buf, parent, f.Name, f, tplFindByEquality)
		case isSortable(f):
			writeFindByTpl(buf, parent, f.Name, f, tplFindByCondition)
		case isCollection(f):
			writeFindByTpl(buf, parent, f.Name, f, tplFindByCollection)
		}
	}
}

func writeFindByTpl(buf *bytes.Buffer, parent *Model, name string, f *Field, tpl string) {
	findableTypeName, ok := f.typeName()
	if !ok {
		return
	}

	query := parent.QueryName
	model := parent.Name
	buf.WriteString(fmt.Sprintf(tpl, name, query, findableTypeName, model))
}

// findableTypeName returns the correct go type name with its qualifier for
// the given type. It returns such name along with a boolean reporting whether
// such type was found or not.
func findableTypeName(typ types.Type, pkg *types.Package) (string, bool) {
	collectionAlreadyScanned := false
	for {
		valid, deepest := lookupValid(pkg, typ)
		if valid != nil {
			return shortName(pkg, typ), true
		}

		if collectionAlreadyScanned {
			break
		}

		singular := collectionElemType(deepest)
		if singular == nil {
			break
		}

		typ = singular
		collectionAlreadyScanned = true
	}

	return "", false
}

func isOneToOneRelationship(f *Field) bool {
	return f.Kind == Relationship && !f.IsOneToManyRelationship()
}

// lookupValid returns the first valid type looking into the underlying types of
// the passed one; if no valid found it will return the deepest underlying
func lookupValid(pkg *types.Package, typ types.Type) (valid, deepest types.Type) {
	if isBasicType(pkg, typ) || isSpecialType(pkg, typ) {
		return typ, nil
	}

	underlying := typ.Underlying()
	switch {
	case typ != underlying:
		return lookupValid(pkg, underlying)
	default:
		return nil, typ
	}
}

// collectionElemType returns the type of the elements of the passed collection,
// or nill if the passed type is not a collection
func collectionElemType(typ types.Type) types.Type {
	switch typ := typ.(type) {
	case *types.Array:
		return typ.Elem()
	case *types.Slice:
		return typ.Elem()
	}

	return nil
}

// isBasicType returns true if passed type is one of the followings:
// string, byte, bool, float(s), int(s) or uint(s)
func isBasicType(pkg *types.Package, typ types.Type) bool {
	switch typ.String() {
	case "string", "bool", "byte", "float64", "float32",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return true
	}

	return false
}

// isSpecialType returns true if the passed type is one of the specialTypes
// or a types.SQLType
func isSpecialType(pkg *types.Package, typ types.Type) bool {
	_, ok := specialTypeShortName(typ)
	return ok || isSQLType(pkg, typ)
}

func specialTypeShortName(typ types.Type) (string, bool) {
	s := removeGoPath(strings.TrimLeft(typ.String(), "*."))
	special, ok := specialTypes[s]
	return special, ok
}

func shortName(pkg *types.Package, typ types.Type) string {
	var prefix string
	if singleType := collectionElemType(typ); singleType != nil {
		t := typ.String()
		idx := strings.Index(t, "]")
		prefix = t[:idx+1]
		typ = singleType
	}

	if specialName, ok := specialTypeShortName(typ); ok {
		return prefix + specialName
	} else {
		shortName := typeString(typ, pkg)
		return prefix + strings.Replace(shortName, "*", "", -1)
	}
}

// isEqualizable returns true if the autogenerated FindBy will use an equal query
func isEqualizable(f *Field) bool {
	return f.Type == "string" || f.Type == "bool" ||
		f.Kind == Interface || f.Type == URL
}

// isSortable returns true if the autogenerated FindBy will use a kallax.ScalarCond
func isSortable(f *Field) bool {
	return f.Kind == Basic && !isEqualizable(f)
}

// isCollection returns true if the autogenerated FindBy will use an kallax.ArrayContains
func isCollection(f *Field) bool {
	return f.Kind == Slice || f.Kind == Array
}
