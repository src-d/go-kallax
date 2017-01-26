package generator

import (
	"bytes"
	"errors"
	"fmt"
	"go/types"
	"reflect"
	"strings"
	"unicode"
)

// https://www.postgresql.org/docs/current/static/sql-keywords-appendix.html
var reservedKeywords = map[string]struct{}{
	"all":               {},
	"analyse":           {},
	"analyze":           {},
	"and":               {},
	"any":               {},
	"array":             {},
	"as":                {},
	"asc":               {},
	"asymmetric":        {},
	"authorization":     {},
	"binary":            {},
	"both":              {},
	"case":              {},
	"cast":              {},
	"check":             {},
	"collate":           {},
	"collation":         {},
	"columns":           {},
	"concurrently":      {},
	"constraint":        {},
	"create":            {},
	"cross":             {},
	"current_catalog":   {},
	"current_date":      {},
	"current_role":      {},
	"current_schema":    {},
	"current_time":      {},
	"current_timestamp": {},
	"current_user":      {},
	"default":           {},
	"deferrable":        {},
	"desc":              {},
	"distinct":          {},
	"do":                {},
	"end":               {},
	"except":            {},
	"false":             {},
	"fetch":             {},
	"for":               {},
	"foreign":           {},
	"from":              {},
	"grant":             {},
	"group":             {},
	"having":            {},
	"in":                {},
	"index":             {},
	"initially":         {},
	"inner":             {},
	"intersect":         {},
	"into":              {},
	"join":              {},
	"key":               {},
	"lateral":           {},
	"leading":           {},
	"left":              {},
	"like":              {},
	"limit":             {},
	"localtime":         {},
	"localtimestamp":    {},
	"natural":           {},
	"not":               {},
	"null":              {},
	"offset":            {},
	"on":                {},
	"only":              {},
	"or":                {},
	"order":             {},
	"overlaps":          {},
	"placing":           {},
	"primary":           {},
	"references":        {},
	"returning":         {},
	"select":            {},
	"some":              {},
	"symmetric":         {},
	"table":             {},
	"then":              {},
	"to":                {},
	"trailing":          {},
	"true":              {},
	"union":             {},
	"unique":            {},
	"user":              {},
	"using":             {},
	"variadic":          {},
	"verbose":           {},
	"when":              {},
	"where":             {},
	"window":            {},
	"with":              {},
}

// special types that are not analyzed because SQL already knows
// how to handle them
var specialTypes = map[string]string{
	"github.com/src-d/go-kallax.ID": "kallax.ID",
	"net/url.URL":                   "url.URL",
	"time.Time":                     "time.Time",
}

// mappings defines the mapping between specific types and their counterpart
// in kallax types
var mappings = map[string]string{
	"url.URL": "types.URL",
}

// Package is the representation of a scanned package.
type Package struct {
	pkg *types.Package
	// Name is the package name.
	Name string
	// Models are all the models found in the package.
	Models []*Model
}

const (
	// StoreNamePattern is the pattern used to name stores.
	StoreNamePattern = "%sStore"
	// QueryNamePattern is the pattern used to name queries.
	QueryNamePattern = "%sQuery"
	// ResultSetNamePattern is the pattern used to name result sets.
	ResultSetNamePattern = "%sResultSet"
)

// Model is the representation of an user-defined model.
type Model struct {
	// Name is the model name.
	Name string
	// StoreName is the name of the store for this model.
	StoreName string
	// QueryName is the name of the query for this model.
	QueryName string
	// ResultSetName is the name of the result set for this model.
	ResultSetName string

	// Table is the name of the table, which will be extracted from the `table`
	// struct tag of the kallax.Model field in the model.
	// If one is not provided, it will be the model name transformed to lower
	// snake case. A model with an empty table name is not valid.
	Table string
	// Type is the string representation of the type.
	Type string
	// Fields contains the list of fields in the model.
	Fields []*Field
	// Events contains the list of events implemented by the model.
	Events Events
	// Node is the node where the model was defined.
	Node *types.Named
	// CtorFunc is a reference to the model constructor.
	CtorFunc *types.Func
	// Package is a reference to the package where the model was defined.
	Package *types.Package
}

// NewModel creates a new model with the given name.
func NewModel(n string) *Model {
	return &Model{
		Name:          n,
		StoreName:     fmt.Sprintf(StoreNamePattern, n),
		QueryName:     fmt.Sprintf(QueryNamePattern, n),
		ResultSetName: fmt.Sprintf(ResultSetNamePattern, n),
		Type:          "struct",
	}
}

// Alias returns the alias of the model, which is the lowercased name preceded
// by "__".
func (m *Model) Alias() string {
	return "__" + strings.ToLower(m.Name)
}

// String prints the representation of the model.
func (m *Model) String() string {
	var events []string
	for _, e := range m.Events {
		events = append(events, string(e))
	}

	return fmt.Sprintf("%q [%d Field(s)] [Events: %s]", m.Name, len(m.Fields), events)
}

// ErrEventConflict is returned whenever the model implements a Save event,
// but also implements an Update or Insert event of the same kind.
var ErrEventConflict = errors.New(
	"kallax: Event conflict a *Save and a *Update or *Insert are present",
)

// repeatedFields returns the list of repeated fields found in the model.
func (m *Model) repeatedFields() []string {
	var occ = make(map[string]uint)
	m.checkFieldOccurrences(m.Fields, occ)

	var names []string
	for name, times := range occ {
		if times > 1 {
			names = append(names, name)
		}
	}
	return names
}

func (m *Model) checkFieldOccurrences(fields []*Field, occurrences map[string]uint) {
	for _, f := range fields {
		if f.Inline() {
			m.checkFieldOccurrences(f.Fields, occurrences)
		} else {
			occurrences[f.Name]++
		}
	}
}

// Validate returns an error if the model is not valid. To be valid, a model
// needs a non-empty table name, a non-repeated set of fields, and no
// conflicting events.
func (m *Model) Validate() error {
	if fields := m.repeatedFields(); len(fields) > 0 {
		return fmt.Errorf("kallax: the following fields are repeated: %v", fields)
	}

	if m.Table == "" {
		return fmt.Errorf("kallax: model %s has no table", m.Name)
	}

	if m.Events.Has(BeforeSave) && m.Events.Has(BeforeInsert) {
		return ErrEventConflict
	}

	if m.Events.Has(BeforeSave) && m.Events.Has(BeforeUpdate) {
		return ErrEventConflict
	}

	if m.Events.Has(AfterSave) && m.Events.Has(AfterInsert) {
		return ErrEventConflict
	}

	if m.Events.Has(AfterSave) && m.Events.Has(AfterUpdate) {
		return ErrEventConflict
	}

	return nil
}

// CtorArgs returns the string with the generated constructor arguments,
// based on the constructor scanned, if any.
func (m *Model) CtorArgs() string {
	if m.CtorFunc == nil {
		return ""
	}

	var ret []string
	sig := m.CtorFunc.Type().(*types.Signature)

	for i := 0; i < sig.Params().Len(); i++ {
		param := sig.Params().At(i)
		typeName := typeString(param.Type(), m.Package)
		paramName := param.Name()
		if paramName == "s" {
			paramName = fmt.Sprintf("arg%v", i)
		}
		ret = append(ret, fmt.Sprintf("%v %v", paramName, typeName))
	}

	return strings.Join(ret, ", ")
}

// CtorArgVars returns the string representation of the variables to call the
// scanned constructor in the generated constructor.
func (m *Model) CtorArgVars() string {
	if m.CtorFunc == nil {
		return ""
	}

	var ret []string
	sig := m.CtorFunc.Type().(*types.Signature)

	for i := 0; i < sig.Params().Len(); i++ {
		ret = append(ret, sig.Params().At(i).Name())
	}

	return strings.Join(ret, ", ")
}

// CtorReturns returns the string representation of the return values of the
// generated constructor based on the ones in the scanned constructor.
func (m *Model) CtorReturns() string {
	if m.CtorFunc == nil {
		return "(record *" + m.Name + ")"
	}

	var ret []string
	hasError := false
	sig := m.CtorFunc.Type().(*types.Signature)

	for i := 0; i < sig.Results().Len(); i++ {
		res := sig.Results().At(i)
		typeName := typeString(res.Type(), m.Package)
		if isTypeOrPtrTo(res.Type(), m.Node) {
			ret = append(ret, "record "+typeName)
		} else if isBuiltinError(res.Type()) && !hasError {
			ret = append(ret, "err "+typeName)
			hasError = true
		} else if res.Name() != "" {
			ret = append(ret, fmt.Sprintf("r%d %v", i, res.Name()))
		} else {
			ret = append(ret, fmt.Sprintf("r%d %v", i, typeName))
		}
	}

	return "(" + strings.Join(ret, ", ") + ")"
}

// CtorRetVars returns the string representation of the return variables to
// receive in the generated constructor based on the ones in the scanned
// constructor.
func (m *Model) CtorRetVars() string {
	if m.CtorFunc == nil {
		return "record"
	}

	var ret []string
	hasError := false
	sig := m.CtorFunc.Type().(*types.Signature)

	for i := 0; i < sig.Results().Len(); i++ {
		res := sig.Results().At(i)
		if isTypeOrPtrTo(res.Type(), m.Node) {
			ret = append(ret, "record")
		} else if isBuiltinError(res.Type()) && !hasError {
			ret = append(ret, "err")
			hasError = true
		} else {
			ret = append(ret, fmt.Sprintf("r%d", i))
		}
	}

	return strings.Join(ret, ", ")
}

// Relationships returns the fields of a model that are relationships.
// NOTE: right now only 1:1 relationships are supported.
func (m *Model) Relationships() []*Field {
	return relationshipsOnFields(m.Fields)
}

func relationshipsOnFields(fields []*Field) []*Field {
	var result []*Field
	for _, f := range fields {
		if f.Kind == Relationship {
			result = append(result, f)
		} else if f.Inline() {
			result = append(result, relationshipsOnFields(f.Fields)...)
		}
	}
	return result
}

// Field is the representation of a model field.
type Field struct {
	// Name is the field name.
	Name string
	// Type is the string representation of the field type.
	Type string
	// Kind is the kind of field.
	Kind FieldKind
	// Node is the reference to the field node.
	Node *types.Var
	// Tag is the strug tag of the field.
	Tag reflect.StructTag
	// Fields contains all the children fields of the field. A field has
	// children only if it is a struct.
	Fields []*Field
	// Parent is a reference to the parent field.
	Parent *Field
	// IsPtr reports whether the field is a pointer type or not.
	IsPtr bool
	// IsJSON reports whether the field has to be converted to JSON.
	IsJSON bool
	// IsAlias reports whether the field is of a type that aliases some other type.
	IsAlias bool
}

// FieldKind is the kind of a field.
type FieldKind int

const (
	// Basic is a field with a basic type.
	Basic FieldKind = iota
	// Array is a field with an array type.
	Array
	// Slice is a field with a slice type.
	Slice
	// Map is a field with a map type.
	Map
	// Interface is a field with an interface type.
	Interface
	// Struct is a field with a struct type.
	Struct
	// Relationship is a field which is a relationship to other model/s.
	Relationship
)

// NewField creates a new field with its name, type and struct tag.
func NewField(n, t string, tag reflect.StructTag) *Field {
	return &Field{
		Name: n,
		Type: t,
		Tag:  tag,
	}
}

// SetFields sets all the children fields and the current field as a parent of
// the children.
func (f *Field) SetFields(sf []*Field) {
	for _, field := range sf {
		field.Parent = f
		f.Fields = append(f.Fields, field)
	}
}

// ColumnName returns the SQL valid column name of the field.
// The struct tag `column` of the field can be use to set the name, otherwise
// is the field name converted to lower snake case.
// If the resultant name is a reserved keyword a _ will be prepended to the name.
func (f *Field) ColumnName() string {
	name := f.Tag.Get("column")
	if name == "" {
		name = toLowerSnakeCase(f.Name)
	}

	if _, ok := reservedKeywords[strings.ToLower(name)]; ok {
		name = "_" + name
	}

	return name
}

// ForeignKey returns the name of the foreign keys as specified in the struct
// tag `fk` or the default foreign key, which is the name of the relationship
// type in lower snake case with "_id" appended.
func (f *Field) ForeignKey() string {
	if f.Kind != Relationship {
		return ""
	}

	fk := f.Tag.Get("fk")
	if fk == "" {
		fk = foreignKeyForType(f.Type)
	}

	return fk
}

// IsOneToManyRelationship returns whether the field is a one to many
// relationship.
func (f *Field) IsOneToManyRelationship() bool {
	return f.Kind == Relationship && strings.HasPrefix(f.Type, "[]")
}

func foreignKeyForType(typ string) string {
	parts := strings.Split(typ, ".")
	typ = parts[len(parts)-1]
	return toLowerSnakeCase(typ) + "_id"
}

// Inline reports whether the field is inline and its children will be in the
// root of the model.
// An inline field is the one having the type kallax.Model or one that has a
// struct tag `kallax` containing `,inline`.
func (f *Field) Inline() bool {
	if f.Type == BaseModel {
		return true
	}

	tag := f.Tag.Get("kallax")
	for _, p := range strings.Split(tag, ",") {
		if p == "inline" {
			return true
		}
	}

	return false
}

func (f *Field) String() string {
	return f.Name
}

func (f *Field) fieldName() string {
	if f.Parent != nil {
		return fmt.Sprintf("%s.%s", f.Parent.fieldName(), f.Name)
	}
	return f.Name
}

func (f *Field) fieldVarName() string {
	return fmt.Sprintf("r.%s", f.fieldName())
}

func (f *Field) fieldVarAddress() string {
	name := f.fieldVarName()
	if !f.IsPtr {
		name = "&" + name
	}

	return name
}

// Address returns the string representation of the code used to get the
// pointer to the field.
func (f *Field) Address() string {
	name := f.fieldVarAddress()
	if mapped, ok := mappings[f.Type]; ok {
		name = fmt.Sprintf("(*%s)(%s)", mapped, name)
	}

	return f.wrapAddress(name)
}

func (f *Field) wrapAddress(ptr string) string {
	if f.IsJSON {
		return fmt.Sprintf("types.JSON(%s), nil", ptr)
	}

	if f.Kind == Slice {
		return fmt.Sprintf("types.Slice(%s), nil", ptr)
	}

	if f.Kind == Array {
		return fmt.Sprintf("types.Array(%s), nil", ptr)
	}

	return fmt.Sprintf("%s, nil", ptr)
}

// Value is the string representation of the code needed to get the value of
// the field in a way that SQL drivers can process.
func (f *Field) Value() string {
	name := f.fieldVarName()

	switch f.Kind {
	case Basic:
		if mapped, ok := mappings[f.Type]; ok {
			name = fmt.Sprintf("(*%s)(%s)", mapped, f.fieldVarAddress())
		}

		if f.IsAlias {
			typ := f.Type
			if f.IsPtr {
				typ = "*" + typ
			}
			return fmt.Sprintf("(%s)(%s), nil", typ, name)
		}
		return name + ", nil"
	case Slice:
		return fmt.Sprintf("types.Slice(%s), nil", name)
	case Array:
		return fmt.Sprintf("types.Array(%s), nil", f.fieldVarAddress())
	}

	if f.IsJSON {
		return fmt.Sprintf("types.JSON(%s), nil", name)
	}

	return name + ", nil"
}

// TypeSchemaName returns the name of the Schema for the field type.
func (f *Field) TypeSchemaName() string {
	parts := strings.Split(f.Type, ".")
	return parts[len(parts)-1]
}

func isTypeOrPtrTo(ptr types.Type, named *types.Named) bool {
	switch ty := ptr.(type) {
	case *types.Pointer:
		if elem, ok := ty.Elem().(*types.Named); ok && elem == named {
			return true
		}
	case *types.Named:
		if ty == named {
			return true
		}
	}
	return false
}

func typeString(ty types.Type, pkg *types.Package) string {
	ret := types.TypeString(ty, types.RelativeTo(pkg))
	parts := strings.Split(ret, "/")
	prefix := ""
	if len(parts) > 1 {
		for _, r := range parts[0] {
			if r == '.' || unicode.IsLetter(r) {
				break
			}
			prefix += string(r)
		}
	}
	return prefix + parts[len(parts)-1]
}

func isBuiltinError(typ types.Type) bool {
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}

	return named.Obj().Name() == "error" && named.Obj().Parent() == types.Universe
}

func toLowerSnakeCase(s string) string {
	var buf bytes.Buffer
	var lastWasUpper bool
	for i, r := range s {
		if unicode.IsUpper(r) && i != 0 && !lastWasUpper {
			buf.WriteRune('_')
		}
		lastWasUpper = unicode.IsUpper(r)
		buf.WriteRune(unicode.ToLower(r))
	}
	return buf.String()
}

// Event is the name of an event.
type Event string

// Events is a collection of events.
type Events []Event

// Has reports whether the event is in the collection.
func (s Events) Has(e Event) bool {
	for _, event := range s {
		if event == e {
			return true
		}
	}

	return false
}

const (
	// BeforeInsert is an event that will happen before Insert opereations.
	// Conflicts with BeforeSave.
	BeforeInsert Event = "BeforeInsert"
	// AfterInsert is an event that will happen after Insert operations.
	// Conflicts with AfterSave.
	AfterInsert Event = "AfterInsert"
	// BeforeUpdate is an event that will happen before Update operations.
	// Conflicts with BeforeSave.
	BeforeUpdate Event = "BeforeUpdate"
	// AfterUpdate is an event that will happen after Update operations.
	// Conflicts with AfterSave.
	AfterUpdate Event = "AfterUpdate"
	// BeforeSave is an event that will happen before Insert or Update
	// operations. Conflicts with BeforeInsert and BeforeUpdate.
	BeforeSave Event = "BeforeSave"
	// AfterSave is an event that will happen after Insert or Update
	// operations. Conflicts with AfterInsert and AfterUpdate.
	AfterSave Event = "AfterSave"
)
