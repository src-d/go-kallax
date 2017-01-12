package generator

import (
	"errors"
	"fmt"
	"go/types"
	"reflect"
	"strings"
	"unicode"
)

var findableTypes = map[string]bool{
	"string":                        true,
	"int":                           true,
	"int8":                          true,
	"int16":                         true,
	"int32":                         true,
	"int64":                         true,
	"uint":                          true,
	"uint8":                         true,
	"uint16":                        true,
	"uint32":                        true,
	"uint64":                        true,
	"float32":                       true,
	"float64":                       true,
	"struct":                        true,
	"bool":                          true,
	"map":                           true,
	"time.Time":                     true,
	"interface{}":                   true,
	"gopkg.in/mgo.v2/bson.ObjectId": true,
}

type Package struct {
	Name      string
	Models    []*Model
	Structs   []string
	Functions []string
}

func (p *Package) StructIsDefined(name string) bool {
	for _, n := range p.Structs {
		if name == n {
			return true
		}
	}

	return false
}

func (p *Package) FunctionIsDefined(name string) bool {
	for _, n := range p.Functions {
		if name == n {
			return true
		}
	}

	return false
}

const (
	StoreNamePattern     = "%sStore"
	QueryNamePattern     = "%sQuery"
	ResultSetNamePattern = "%sResultSet"
)

type Model struct {
	Name          string
	StoreName     string
	QueryName     string
	ResultSetName string

	Collection  string
	Type        string
	Fields      []*Field
	New         bool
	Init        bool
	Events      Events
	CheckedNode *types.Named
	NewFunc     *types.Func
	Package     *types.Package
}

func NewModel(n string) *Model {
	return &Model{
		Name:          n,
		StoreName:     fmt.Sprintf(StoreNamePattern, n),
		QueryName:     fmt.Sprintf(QueryNamePattern, n),
		ResultSetName: fmt.Sprintf(ResultSetNamePattern, n),
		Type:          "struct",
		Fields:        make([]*Field, 0),
		Events:        make([]Event, 0),
	}
}

func (m *Model) String() string {
	var events []string
	for _, e := range m.Events {
		events = append(events, string(e))
	}

	return fmt.Sprintf("%q [%d Field(s)] [Events: %s]", m.Name, len(m.Fields), events)
}

func (m *Model) ValidFields() []*Field {
	var fields []*Field
	for _, f := range m.Fields {
		if f.Findable() {
			fields = append(fields, f)
		}
	}

	return fields
}

var (
	ErrEventConflict = errors.New(
		"Event conflict a *Save and a *Update or *Insert are present",
	)
)

func (m *Model) Validate() error {
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

func (m *Model) NewArgs() string {
	if m.NewFunc == nil {
		return ""
	}

	var ret []string
	sig := m.NewFunc.Type().(*types.Signature)

	for i := 0; i < sig.Params().Len(); i++ {
		param := sig.Params().At(i)

		if m.isStore(param.Type()) {
			continue
		}

		typeName := typeString(param.Type(), m.Package)
		paramName := param.Name()
		if paramName == "s" {
			paramName = fmt.Sprintf("arg%v", i)
		}
		ret = append(ret, fmt.Sprintf("%v %v", paramName, typeName))
	}

	return strings.Join(ret, ", ")
}

func (m *Model) NewArgVars() string {
	if m.NewFunc == nil {
		return ""
	}

	var ret []string
	sig := m.NewFunc.Type().(*types.Signature)

	for i := 0; i < sig.Params().Len(); i++ {
		param := sig.Params().At(i)

		if m.isStore(param.Type()) {
			ret = append(ret, "s")
			continue
		}

		paramName := param.Name()
		if paramName == "s" {
			paramName = fmt.Sprintf("arg%v", i)
		}
		ret = append(ret, paramName)
	}

	return strings.Join(ret, ", ")
}

func (m *Model) NewReturns() string {
	if m.NewFunc == nil {
		return "(doc *" + m.Name + ")"
	}

	var ret []string
	hasError := false
	sig := m.NewFunc.Type().(*types.Signature)

	for i := 0; i < sig.Results().Len(); i++ {
		res := sig.Results().At(i)
		typeName := typeString(res.Type(), m.Package)
		if isTypeOrPtrTo(res.Type(), m.CheckedNode) {
			ret = append(ret, "doc "+typeName)
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

func (m *Model) NewRetVars() string {
	if m.NewFunc == nil {
		return "doc"
	}

	var ret []string
	hasError := false
	sig := m.NewFunc.Type().(*types.Signature)

	for i := 0; i < sig.Results().Len(); i++ {
		res := sig.Results().At(i)
		if isTypeOrPtrTo(res.Type(), m.CheckedNode) {
			ret = append(ret, "doc")
		} else if isBuiltinError(res.Type()) && !hasError {
			ret = append(ret, "err")
			hasError = true
		} else {
			ret = append(ret, fmt.Sprintf("r%d", i))
		}
	}

	return strings.Join(ret, ", ")
}

func (m *Model) isStore(typ types.Type) bool {
	if isPtrToInvalid(typ) {
		return true
	}
	if ptrTo, ok := typ.(*types.Pointer); ok {
		if named, ok := ptrTo.Elem().(*types.Named); ok && named.Obj().Name() == m.StoreName {
			return true
		}
	}
	return false
}

type Function struct {
	Name string
	Args string
}

func NewFunction() {
}

type Field struct {
	Name        string
	Type        string
	CheckedNode *types.Var
	Tag         reflect.StructTag
	Fields      []*Field
	Parent      *Field
	isMap       bool
}

func NewField(n, t string, tag reflect.StructTag) *Field {
	return &Field{
		Name:   n,
		Type:   t,
		Tag:    tag,
		Fields: make([]*Field, 0),
		isMap:  strings.HasPrefix(t, "map["),
	}
}

func (f *Field) SetFields(sf []*Field) {
	for _, field := range sf {
		f.AddField(field)
	}
}

func (f *Field) AddField(field *Field) {
	field.Parent = f
	f.Fields = append(f.Fields, field)
}

func (f *Field) GetPath() string {
	recursive := f
	path := make([]string, 0)
	done := map[*Field]bool{}
	for recursive != nil {
		if recursive.isMap {
			path = append(path, "[map]")
		}

		if !recursive.Inline() {
			path = append(path, recursive.DbName())
		}

		recursive = recursive.Parent
		if done[recursive] {
			break
		}
		done[recursive] = true
	}

	return strings.Join(reverseSliceStrings(path), ".")
}

func (f *Field) ContainsMap() bool {
	return f.containsMap(map[*Field]bool{})
}

func (f *Field) containsMap(checked map[*Field]bool) bool {
	if checked[f] {
		return false
	}
	checked[f] = true

	if !f.isMap && f.Parent != nil {
		return f.Parent.containsMap(checked)
	}

	return f.isMap
}

func (f *Field) GetTagValue(key string) string {
	if f.Tag == "" {
		return ""
	}

	return f.Tag.Get(key)
}

func (f *Field) DbName() string {
	name := f.GetTagValue("bson")
	endFieldName := strings.Index(name, ",")
	if endFieldName != -1 {
		name = name[:endFieldName]
	}

	if name == "" {
		name = strings.ToLower(f.Name)
	}

	return name
}

func (f *Field) Inline() bool {
	tag := f.GetTagValue("bson")
	for _, p := range strings.Split(tag, ",") {
		if p == "inline" {
			return true
		}
	}

	return false
}

func (f *Field) ValidFields() []*Field {
	fields := make([]*Field, 0)
	for _, f := range f.Fields {
		if f.Findable() {
			fields = append(fields, f)
		}
	}

	return fields
}

func (f *Field) FindableType() string {
	startType := strings.Index(f.Type, "]")
	if startType != -1 {
		return f.Type[startType+1:]
	}

	return f.Type
}

func (f *Field) Findable() bool {
	return findableTypes[f.FindableType()]
}

func (f *Field) String() string {
	return f.Name
}

func reverseSliceStrings(input []string) []string {
	if len(input) == 0 {
		return input
	}

	return append(reverseSliceStrings(input[1:]), input[0])
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

func isPtrToInvalid(ty types.Type) bool {
	ptrTo, ok := ty.(*types.Pointer)
	return ok && ptrTo.Elem() == types.Typ[types.Invalid]
}

func isBuiltinError(typ types.Type) bool {
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}

	return named.Obj().Name() == "error" && named.Obj().Parent() == types.Universe
}

type Event string

type Events []Event

func (s Events) Has(e Event) bool {
	for _, event := range s {
		if event == e {
			return true
		}
	}

	return false
}

const (
	BeforeInsert Event = "BeforeInsert"
	AfterInsert  Event = "AfterInsert"
	BeforeUpdate Event = "BeforeUpdate"
	AfterUpdate  Event = "AfterUpdate"
	BeforeSave   Event = "BeforeSave"
	AfterSave    Event = "AfterSave"
)
