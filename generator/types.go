package generator

import (
	"bytes"
	"fmt"
	"go/types"
	"reflect"
	"strconv"
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
	"github.com/networkteam/go-kallax.UUID":      "kallax.UUID",
	"github.com/networkteam/go-kallax.ULID":      "kallax.ULID",
	"github.com/networkteam/go-kallax.NumericID": "kallax.NumericID",
	"github.com/satori/go.uuid.UUID":             "kallax.UUID",
	"github.com/gofrs/uuid.UUID":                 "kallax.UUID",
	"net/url.URL":                                "url.URL",
	"time.Time":                                  "time.Time",
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
	Models        []*Model
	indexedModels map[string]*Model
}

// NewPackage creates a new package.
func NewPackage(pkg *types.Package) *Package {
	return &Package{
		Name:          pkg.Name(),
		pkg:           pkg,
		indexedModels: make(map[string]*Model),
	}
}

// SetModels sets the models of the packages and indexes them.
func (p *Package) SetModels(models []*Model) {
	for _, m := range models {
		p.indexedModels[m.Name] = m
	}
	p.Models = models
}

// FindModel finds the model with the given name.
func (p *Package) FindModel(name string) *Model {
	return p.indexedModels[name]
}

func (p *Package) addMissingRelationships() error {
	for _, m := range p.Models {
		for _, f := range m.Fields {
			if f.Kind == Relationship && !f.IsInverse() {
				if err := p.trySetFK(f.TypeSchemaName(), f); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (p *Package) trySetFK(model string, fk *Field) error {
	m := p.FindModel(model)
	if m == nil {
		return fmt.Errorf("kallax: cannot assign implicit foreign key to non-existent model %s", model)
	}

	var found bool
	for _, f := range m.Fields {
		if f.Kind == Relationship {
			if f.ForeignKey() == fk.ForeignKey() {
				found = true
				break
			}
		} else {
			if f.ColumnName() == fk.ForeignKey() {
				found = true
				break
			}
		}
	}

	if !found {
		for _, ifk := range m.ImplicitFKs {
			if ifk.Name == fk.ForeignKey() {
				found = true
				break
			}
		}
	}

	if !found {
		m.ImplicitFKs = append(m.ImplicitFKs, ImplicitFK{
			Name: fk.ForeignKey(),
			Type: identifierType(fk.Model.ID),
		})
	}
	return nil
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
	// ImplicitFKs contains the list of fks that are implicit based on
	// other models' definitions, such as foreign keys with no explicit inverse
	// on the related model.
	ImplicitFKs []ImplicitFK
	// ID contains the identifier field of the model.
	ID *Field
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

	return fmt.Sprintf("%q [%d Field(s)] [Events: %s]", m.Name, len(m.Fields)-1, events)
}

type occurrences map[string]uint

func (o occurrences) inc(name string) {
	o[name]++
}

func (o occurrences) repeated() []string {
	var result []string
	for v, times := range o {
		if times > 1 {
			result = append(result, v)
		}
	}
	return result
}

// repeatedFields returns the list of repeated fields found in the model.
func (m *Model) repeatedFields() []string {
	var occ = make(occurrences)
	m.checkFieldOccurrences(m.Fields, occ)
	return occ.repeated()
}

func (m *Model) checkFieldOccurrences(fields []*Field, occurrences occurrences) {
	for _, f := range fields {
		if f.Inline() {
			m.checkFieldOccurrences(f.Fields, occurrences)
		} else {
			occurrences.inc(f.Name)
		}
	}
}

func (m *Model) repeatedCols() []string {
	columns := make(occurrences)
	m.checkFieldColumns(m.Fields, columns)
	return columns.repeated()
}

func (m *Model) checkFieldColumns(fields []*Field, cols occurrences) {
	for _, f := range fields {
		if f.Inline() {
			m.checkFieldColumns(f.Fields, cols)
		} else if f.Kind != Relationship {
			cols.inc(f.ColumnName())
		}
	}
}

// Validate returns an error if the model is not valid. To be valid, a model
// needs a non-empty table name, a non-repeated set of fields.
func (m *Model) Validate() error {
	if m.ID == nil {
		return fmt.Errorf("kallax: model %s has no primary key defined", m.Name)
	}

	if !isValidIdentifier(m.ID) {
		return fmt.Errorf("kallax: primary key %q of model %q does not have a valid identifier type (%s)", m.ID.Name, m.Name, m.ID.Type)
	}

	if fields := m.repeatedFields(); len(fields) > 0 {
		return fmt.Errorf("kallax: the following fields are repeated: %v", fields)
	}

	if cols := m.repeatedCols(); len(cols) > 0 {
		return fmt.Errorf("kallax: the following column names are repeated: %v", cols)
	}

	if m.Table == "" {
		return fmt.Errorf("kallax: model %s has no table", m.Name)
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

	paramsLen := sig.Params().Len()
	for i := 0; i < paramsLen; i++ {
		param := sig.Params().At(i)

		// TODO: refactor findableTypeName so this is not needed
		// or split into two functions
		typeName, ok := findableTypeName(param.Type(), m.Package)
		if !ok {
			typeName = typeString(param.Type(), m.Package)
		}

		if paramsLen == i+1 && sig.Variadic() {
			typeName = "..." + typeName
		} else // TODO: Dirty fix for #229, address properly inside findableTypeName or typeString
		if collectionElemType(param.Type()) != nil && !strings.HasPrefix(typeName, "[]") {
			typeName = "[]" + typeName
		}

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

	paramsLen := sig.Params().Len()
	for i := 0; i < sig.Params().Len(); i++ {
		arg := sig.Params().At(i).Name()
		if paramsLen == i+1 && sig.Variadic() {
			arg += "..."
		}
		ret = append(ret, arg)
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
		// TODO: refactor findableTypeName so this is not needed
		// or split into two functions
		typeName, ok := findableTypeName(res.Type(), m.Package)
		if !ok {
			typeName = typeString(res.Type(), m.Package)
		}
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

// SetFields sets all the children fields and their model to the current
// model.
// It also finds the primary key and sets it in the model.
// It will return an error if more than one primary key is found.
// SetFields always sets the primary key as the first field of the model.
// So, all models can expect to have the primary key in the position 0 of
// their field slice. This is because the Store will expect the ID in that
// position.
func (m *Model) SetFields(fields []*Field) error {
	var fs []*Field
	var id *Field
	for _, f := range flattenFields(fields) {
		f.Model = m
		if f.IsPrimaryKey() && f.Type != BaseModel {
			if id != nil {
				return fmt.Errorf(
					"kallax: found more than one primary key in model %s: %s and %s",
					m.Name,
					id.Name,
					f.Name,
				)
			}

			id = f
		} else if f.IsPrimaryKey() {
			if f.primaryKey == "" {
				return fmt.Errorf(
					"kallax: primary key defined in %s has no field name, but it must be specified",
					f.Name,
				)
			}

			// the pk is defined in the model, we need to collect the model
			// and we'll look for the field afterwards, when we have collected
			// all fields. The model is appended to the field set, though,
			// because it will not act as a primary key.
			id = f
			fs = append(fs, f)
		} else {
			fs = append(fs, f)
		}
	}

	// if the id is a Model we need to look for the specified field
	if id != nil && id.Type == BaseModel {
		for i, f := range fs {
			if f.columnName == id.primaryKey {
				f.isPrimaryKey = true
				f.isAutoincrement = id.isAutoincrement
				id = f

				if len(fs)-1 == i {
					fs = append(fs[:i])
				} else {
					fs = append(fs[:i], fs[i+1:]...)
				}
				break
			}
		}

		// If the ID is still a base model, means we did not find the pk
		// field.
		if id.Type == BaseModel {
			return fmt.Errorf(
				"kallax: the primary key was supposed to be %s according to the pk definition in %s, but the field could not be found",
				id.primaryKey,
				id.Name,
			)
		}
	}

	if id != nil {
		m.Fields = []*Field{id}
		m.ID = id
	}
	m.Fields = append(m.Fields, fs...)
	return nil
}

// Relationships returns the fields of a model that are relationships.
func (m *Model) Relationships() []*Field {
	return relationshipsOnFields(m.Fields)
}

// Inverses returns the inverse relationships of the model.
func (m *Model) Inverses() []*Field {
	var inverses []*Field
	for _, f := range relationshipsOnFields(m.Fields) {
		if f.IsInverse() {
			inverses = append(inverses, f)
		}
	}
	return inverses
}

// NonInverses returns the relationships of the model that are not inverses.
func (m *Model) NonInverses() []*Field {
	var rels []*Field
	for _, f := range relationshipsOnFields(m.Fields) {
		if !f.IsInverse() {
			rels = append(rels, f)
		}
	}
	return rels
}

// HasRelationships returns whether the model has relationships or not.
func (m *Model) HasRelationships() bool {
	return len(m.Relationships()) > 0
}

// HasInverses returns whether the model has inverse relationships or not.
func (m *Model) HasInverses() bool {
	return len(m.Inverses()) > 0
}

// HasNonInverses returns whether the model has non inverse relationships or not.
func (m *Model) HasNonInverses() bool {
	return len(m.NonInverses()) > 0
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

// ImplicitFK is a foreign key that is defined on just one side of the
// relationship and needs to be added on the other side.
type ImplicitFK struct {
	Name string
	Type string
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
	// Model is the reference to the model containing this field.
	Model *Model
	// IsPtr reports whether the field is a pointer type or not.
	IsPtr bool
	// IsJSON reports whether the field has to be converted to JSON.
	IsJSON bool
	// IsAlias reports whether the field is of a type that aliases some other type.
	IsAlias bool
	// IsEmbedded reports whether the field is an embedded struct or not.
	// A struct is considered embedded if and only if the struct was embedded
	// as defined in Go.
	IsEmbedded bool

	primaryKey      string
	isPrimaryKey    bool
	isUnique        bool
	isAutoincrement bool
	columnName      string
}

// FieldKind is the kind of a field.
type FieldKind int

const (
	// Basic is a field with a basic type.
	// On top of the Go basic types, we consider Basic as well the following
	// types:
	//  - time.Time
	//  - time.Duration
	//  - url.URL
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
	// Invalid is an invalid field type.
	Invalid
)

// String returns the constant name of the FieldKind
func (t FieldKind) String() string {
	switch t {
	case Basic:
		return "Basic"
	case Array:
		return "Array"
	case Slice:
		return "Slice"
	case Map:
		return "Map"
	case Interface:
		return "Interface"
	case Struct:
		return "Struct"
	case Relationship:
		return "Relationship"
	case Invalid:
		return "Invalid"
	default:
		return "UNKNOWN"
	}
}

// NewField creates a new field with its name, type and struct tag.
func NewField(n, t string, tag reflect.StructTag) *Field {
	pkName, autoincr, isPrimaryKey := pkProperties(tag)

	return &Field{
		Name: n,
		Type: t,
		Tag:  tag,

		primaryKey:      pkName,
		columnName:      columnName(n, tag),
		isPrimaryKey:    isPrimaryKey,
		isUnique:        isUnique(tag),
		isAutoincrement: autoincr,
	}
}

func isUnique(tag reflect.StructTag) bool {
	return tag.Get("unique") == "true"
}

// pkProperties returns the primary key properties from a struct tag.
// Valid primary key definitions are the following:
// - pk:"" -> non-autoincr primary key without a field name.
// - pk:"autoincr" -> autoincr primary key without a field name.
// - pk:"foobar" -> non-autoincr primary key with a field name.
// - pk:"foobar,autoincr" -> autoincr primary key with a field name.
func pkProperties(tag reflect.StructTag) (name string, autoincr, isPrimaryKey bool) {
	val, ok := tag.Lookup("pk")
	if !ok {
		return
	}

	isPrimaryKey = true
	if val == "autoincr" || val == "" {
		if val == "autoincr" {
			autoincr = true
		}
		return
	}

	parts := strings.Split(val, ",")
	name = parts[0]
	if len(parts) > 1 && parts[1] == "autoincr" {
		autoincr = true
	}

	return
}

// SetFields sets all the children fields and the current field as a parent of
// the children.
func (f *Field) SetFields(sf []*Field) {
	for _, field := range sf {
		field.Parent = f
		field.Model = f.Model
		f.Fields = append(f.Fields, field)
	}
}

// ColumnName returns the SQL valid column name of the field.
// The struct tag `kallax` of the field can be use to set the name, otherwise
// is the field name converted to lower snake case.
// If the resultant name is a reserved keyword a _ will be prepended to the name.
func (f *Field) ColumnName() string {
	return f.columnName
}

func columnName(name string, tag reflect.StructTag) string {
	n := strings.TrimSpace(strings.Split(tag.Get("kallax"), ",")[0])
	if n == "" {
		n = toLowerSnakeCase(name)
	}

	if _, ok := reservedKeywords[strings.ToLower(n)]; ok {
		n = "_" + n
	}

	return n
}

// ForeignKey returns the name of the foreign keys as specified in the struct
// tag `fk` or the default foreign key, which is the name of the relationship
// type in lower snake case with "_id" appended.
func (f *Field) ForeignKey() string {
	if f.Kind != Relationship {
		return ""
	}

	fk := strings.Split(f.Tag.Get("fk"), ",")[0]
	if fk == "" && !f.IsInverse() {
		fk = foreignKeyForModel(f.Model.Name)
	} else if fk == "" {
		fk = foreignKeyForModel(f.TypeSchemaName())
	}

	return fk
}

// IsPrimaryKey reports whether the field is the primary key.
func (f *Field) IsPrimaryKey() bool {
	return f.isPrimaryKey
}

// IsUnique reports whether the field is unique.
func (f *Field) IsUnique() bool {
	return f.isUnique
}

// IsAutoIncrement reports whether the field is an autoincrementable primary key.
func (f *Field) IsAutoIncrement() bool {
	return f.isAutoincrement
}

// IsInverse returns whether the field is an inverse relationship.
func (f *Field) IsInverse() bool {
	if f.Kind != Relationship {
		return false
	}

	for _, part := range strings.Split(f.Tag.Get("fk"), ",") {
		if part == "inverse" {
			return true
		}
	}

	return false
}

// IsOneToManyRelationship returns whether the field is a one to many
// relationship.
func (f *Field) IsOneToManyRelationship() bool {
	return f.Kind == Relationship && strings.HasPrefix(f.Type, "[]")
}

func foreignKeyForModel(model string) string {
	return toLowerSnakeCase(model) + "_id"
}

// Inline reports whether the field is inline and its children will be in the
// root of the model.
// An inline field is the one having the type kallax.Model, one that has a
// struct tag `kallax` containing `,inline` or an embedded struct field.
func (f *Field) Inline() bool {
	if f.Type == BaseModel || f.IsEmbedded {
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

// JSONName returns the name of the field or its JSON name specified in the
// JSON struct tag.
func (f *Field) JSONName() string {
	tag := strings.Split(f.Tag.Get("json"), ",")[0]
	if tag == "" {
		tag = f.Name
	}
	return tag
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
	var casted bool
	if mapped, ok := mappings[f.Type]; ok {
		name = fmt.Sprintf("(*%s)(%s)", mapped, name)
		casted = true
	}

	return f.wrapAddress(name, casted)
}

func (f *Field) typeName() (string, bool) {
	return findableTypeName(f.Node.Type(), f.Node.Pkg())
}

func (f *Field) wrapAddress(ptr string, casted bool) string {
	if f.IsJSON {
		return fmt.Sprintf("types.JSON(%s)", ptr)
	}

	if f.Kind == Slice {
		if typ, ok := castSlice(f); ok {
			return fmt.Sprintf("types.Slice((*%s)(%s))", typ, ptr)
		}
		return fmt.Sprintf("types.Slice(%s)", ptr)
	}

	if f.Kind == Array {
		return fmt.Sprintf("types.Array(%s, %d)", ptr, arrayLen(f))
	}

	if f.IsPtr && !casted {
		if f.Kind == Interface {
			return fmt.Sprintf("types.Nullable(%s)", ptr)
		}
		return fmt.Sprintf("types.Nullable(&%s)", ptr)
	}

	return ptr
}

// Value is the string representation of the code needed to get the value of
// the field in a way that SQL drivers can process.
func (f *Field) Value() string {
	name := f.fieldVarName()

	if f.IsJSON {
		return fmt.Sprintf("types.JSON(%s), nil", name)
	}

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
		return fmt.Sprintf("types.Array(%s, %d), nil", f.fieldVarAddress(), arrayLen(f))
	}

	return name + ", nil"
}

// TypeSchemaName returns the name of the Schema for the field type.
func (f *Field) TypeSchemaName() string {
	parts := strings.Split(f.Type, ".")
	return parts[len(parts)-1]
}

func (f *Field) SQLType() string {
	return f.Tag.Get("sqltype")
}

var identifierTypes = map[string]string{
	"github.com/networkteam/go-kallax.UUID":      "kallax.UUID",
	"github.com/networkteam/go-kallax.ULID":      "kallax.ULID",
	"github.com/networkteam/go-kallax.NumericID": "kallax.NumericID",
	"github.com/satori/go.uuid.UUID":             "kallax.UUID",
	"github.com/gofrs/uuid.UUID":                 "kallax.UUID",
	"int64":                                      "kallax.NumericID",
}

func identifierType(f *Field) string {
	return identifierTypes[typeName(f.Node.Type())]
}

func isValidIdentifier(f *Field) bool {
	_, ok := identifierTypes[typeName(f.Node.Type())]
	return ok
}

func arrayLen(f *Field) int {
	if f.Kind != Array {
		return 0
	}

	idx := strings.Index(f.Type, "]")
	len, _ := strconv.Atoi(f.Type[1:idx])
	return len
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

var supportedSliceTypes = map[string]struct{}{
	"int8":          {},
	"uint8":         {},
	"byte":          {},
	"int16":         {},
	"uint16":        {},
	"int32":         {},
	"uint32":        {},
	"int":           {},
	"uint":          {},
	"int64":         {},
	"uint64":        {},
	"float32":       {},
	"float64":       {},
	"bool":          {},
	"string":        {},
	"time.Time":     {},
	"time.Duration": {},
	"net/url.URL":   {},
}

// castSlice returns the type to which the slice has to be casted and a bool
// reporting whether the slice can have/needs a casting.
// A slice only needs a cast if the type is an alias and the slice underlying
// type is included in `supportedSliceTypes`.
func castSlice(f *Field) (cast string, ok bool) {
	if !strings.HasPrefix(f.Type, "[]") {
		return
	}

	if f.Node == nil {
		return
	}

	if _, isNamed := f.Node.Type().(*types.Named); !isNamed {
		return
	}

	prefix := "[]"
	typ := f.Type[2:]
	if strings.HasPrefix(typ, "*") {
		prefix += "*"
		typ = typ[1:]
	}

	cast = typ
	if idx := strings.LastIndex(cast, "/"); idx >= 0 {
		cast = cast[idx+1:]
	}

	_, ok = supportedSliceTypes[typ]
	return prefix + cast, ok
}

func typeString(ty types.Type, pkg *types.Package) string {
	ret := types.TypeString(ty, types.RelativeTo(pkg))
	parts := strings.Split(ret, string(separator))
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

// flattenFields will recursively flatten all fields removing the embedded ones
// from the field set.
func flattenFields(fields []*Field) []*Field {
	var result = make([]*Field, 0, len(fields))

	for _, f := range fields {
		if f.IsEmbedded && f.Type != BaseModel {
			result = append(result, flattenFields(f.Fields)...)
		} else {
			result = append(result, f)
		}
	}

	return result
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
	// BeforeInsert is an event that will happen before Insert operations.
	BeforeInsert Event = "BeforeInsert"
	// AfterInsert is an event that will happen after Insert operations.
	AfterInsert Event = "AfterInsert"
	// BeforeUpdate is an event that will happen before Update operations.
	BeforeUpdate Event = "BeforeUpdate"
	// AfterUpdate is an event that will happen after Update operations.
	AfterUpdate Event = "AfterUpdate"
	// BeforeSave is an event that will happen before Insert or Update
	// operations.
	BeforeSave Event = "BeforeSave"
	// AfterSave is an event that will happen after Insert or Update
	// operations.
	AfterSave Event = "AfterSave"
	// BeforeDelete is an event that will happen before Delete.
	BeforeDelete Event = "BeforeDelete"
	// AfterDelete is an event that will happen after Delete.
	AfterDelete Event = "AfterDelete"
)
