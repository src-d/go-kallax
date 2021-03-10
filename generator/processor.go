package generator

import (
	"fmt"
	"go/build"
	"go/types"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	// BaseModel is the type name of the kallax base model.
	BaseModel = "github.com/loyalguru/go-kallax.Model"
	//URL is the type name of the net/url.URL.
	URL = "url.URL"
)

type goPath []string

var defaultGoPath = goPath(filepath.SplitList(build.Default.GOPATH))

// Processor is in charge of processing the package in a patch and
// scan models from it.
type Processor struct {
	// Path of the package.
	Path string
	// Ignore is the list of files to ignore when scanning.
	Ignore map[string]struct{}
	// Package is the scanned package.
	Package *types.Package
	silent  bool
}

// NewProcessor creates a new Processor for the given path and ignored files.
func NewProcessor(path string, ignore []string) *Processor {
	i := make(map[string]struct{})
	for _, file := range ignore {
		i[file] = struct{}{}
	}

	return &Processor{
		Path:   path,
		Ignore: i,
	}
}

// Silent makes the processor not spit any output to stdout.
func (p *Processor) Silent() {
	p.silent = true
}

func (p *Processor) write(msg string, args ...interface{}) {
	if !p.silent {
		fmt.Println(fmt.Sprintf(msg, args...))
	}
}

// Do performs all the processing and returns the scanned package.
func (p *Processor) Do() (*Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Dir:  ".",
		Mode: packages.NeedImports | packages.NeedTypes | packages.NeedDeps,
	}, p.Path)
	if err != nil {
		return nil, err
	}
	p.Package = pkgs[0].Types

	return p.processPackage()
}

func (p *Processor) processPackage() (*Package, error) {
	pkg := NewPackage(p.Package)
	var ctors []*types.Func

	p.write("Package: %s", pkg.Name)

	s := p.Package.Scope()
	var models []*Model
	for _, name := range s.Names() {
		obj := s.Lookup(name)
		switch t := obj.Type().(type) {
		case *types.Signature:
			if strings.HasPrefix(name, "new") {
				ctors = append(ctors, obj.(*types.Func))
			}
		case *types.Named:
			if str, ok := t.Underlying().(*types.Struct); ok {
				if m, err := p.processModel(name, str, t); err != nil {
					return nil, err
				} else if m != nil {
					p.write("Model: %s", m)

					if err := m.Validate(); err != nil {
						return nil, err
					}

					models = append(models, m)
					m.Node = t
					m.Package = p.Package
				}
			}
		}
	}

	pkg.SetModels(models)
	if err := pkg.addMissingRelationships(); err != nil {
		return nil, err
	}
	for _, ctor := range ctors {
		p.tryMatchConstructor(pkg, ctor)
	}

	return pkg, nil
}

func (p *Processor) tryMatchConstructor(pkg *Package, fun *types.Func) {
	if !strings.HasPrefix(fun.Name(), "new") {
		return
	}

	if m := pkg.FindModel(fun.Name()[3:]); m != nil {
		sig := fun.Type().(*types.Signature)
		if sig.Recv() != nil {
			return
		}

		res := sig.Results()
		if res.Len() > 0 {
			for i := 0; i < res.Len(); i++ {
				if isTypeOrPtrTo(res.At(i).Type(), m.Node) {
					m.CtorFunc = fun
					return
				}
			}
		}
	}
}

func (p *Processor) processModel(name string, s *types.Struct, t *types.Named) (*Model, error) {
	m := NewModel(name)
	m.Events = p.findEvents(t)

	var base int
	var fields []*Field
	if base, fields = p.processFields(s, nil, true); base == -1 {
		return nil, nil
	}

	p.processBaseField(m, fields[base])
	if err := m.SetFields(fields); err != nil {
		return nil, err
	}

	return m, nil
}

var allEvents = Events{
	BeforeInsert,
	AfterInsert,
	BeforeUpdate,
	AfterUpdate,
	BeforeSave,
	AfterSave,
	BeforeDelete,
	AfterDelete,
}

func (p *Processor) findEvents(node *types.Named) []Event {
	var events []Event
	for _, e := range allEvents {
		if p.isEventPresent(node, e) {
			events = append(events, e)
		}
	}

	return events
}

// isEventPresent checks the given Event is implemented for the given node.
func (p *Processor) isEventPresent(node *types.Named, e Event) bool {
	signature := getMethodSignature(p.Package, types.NewPointer(node), string(e))
	return signatureMatches(signature, nil, typeCheckers{isBuiltinError})
}

// processFields returns which field index is an embedded kallax.Model, or -1 if none.
func (p *Processor) processFields(s *types.Struct, done []*types.Struct, root bool) (base int, fields []*Field) {
	base = -1

	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		if !f.Exported() || isIgnoredField(s, i) {
			continue
		}

		field := NewField(
			f.Name(),
			typeName(f.Type().Underlying()),
			reflect.StructTag(s.Tag(i)),
		)
		field.Node = f
		if typeName(f.Type()) == BaseModel {
			base = i
			field.Type = BaseModel
		}

		if f.Anonymous() {
			field.IsEmbedded = true
		}

		p.processField(field, f.Type(), done, root)
		if field.Kind == Invalid {
			p.write("WARNING: arrays of relationships are not supported. Field %s will be ignored.", field.Name)
			continue
		}

		fields = append(fields, field)
	}

	return base, fields
}

// processField processes recursively the field. During the processing several
// field properties might be modified, such as the properties that report if
// the type has to be serialized to json, if it's an alias or if it's a pointer
// and so on. Also, the kind of the field is set here.
// If root is true, models are established as relationships. If not, they are
// just treated as structs.
// The following types are always set as JSON:
//  - Map
//  - Slice or Array with non-basic underlying type
//  - Interface
//  - Struct that is not a model or is not at root level
func (p *Processor) processField(field *Field, typ types.Type, done []*types.Struct, root bool) {
	switch typ := typ.(type) {
	case *types.Basic:
		field.Type = typ.String()
		field.Kind = Basic
	case *types.Pointer:
		field.IsPtr = true
		p.processField(field, typ.Elem(), done, root)
	case *types.Named:
		if field.Type == BaseModel {
			p.processField(field, typ.Underlying(), done, root)
			return
		}

		if isModel(typ, true) && root {
			field.Kind = Relationship
			field.Type = typ.String()
			return
		}

		// embedded fields won't be stored, only their fields, so it's irrelevant
		// if they implement scanner and valuer
		if !field.IsEmbedded && isSQLType(p.Package, types.NewPointer(typ)) {
			field.Kind = Interface
			return
		}

		if t, ok := specialTypes[typeName(typ)]; ok {
			field.Type = t
			return
		}

		p.processField(field, typ.Underlying(), done, root)
		field.IsAlias = !field.IsJSON
	case *types.Array:
		var underlying Field
		p.processField(&underlying, typ.Elem(), done, root)
		if underlying.Kind == Relationship {
			field.Kind = Invalid
			return
		}

		if underlying.Kind != Basic {
			field.IsJSON = true
		}
		field.Kind = Array
		field.SetFields(underlying.Fields)
	case *types.Slice:
		var underlying Field
		p.processField(&underlying, typ.Elem(), done, root)
		if underlying.Kind == Relationship {
			field.Kind = Relationship
			return
		}

		if underlying.Kind != Basic {
			field.IsJSON = true
		}
		field.Kind = Slice
		field.SetFields(underlying.Fields)
	case *types.Map:
		field.Kind = Map
		field.IsJSON = true
	case *types.Interface:
		field.Kind = Interface
		field.IsJSON = true
	case *types.Struct:
		field.Kind = Struct
		field.IsJSON = true

		d := false
		for _, v := range done {
			if v == typ {
				d = true
				break
			}
		}

		if !d {
			_, subfs := p.processFields(typ, append(done, typ), false)
			field.SetFields(subfs)
		}
	default:
		p.write("WARNING: Ignored field %s of type %s.", field.Name, field.Type)
	}
}

func isSQLType(pkg *types.Package, typ types.Type) bool {
	scan := getMethodSignature(pkg, typ, "Scan")
	if !signatureMatches(scan, typeCheckers{isEmptyInterface}, typeCheckers{isBuiltinError}) {
		return false
	}

	value := getMethodSignature(pkg, typ, "Value")
	if !signatureMatches(value, nil, typeCheckers{isDriverValue, isBuiltinError}) {
		return false
	}

	return true
}

func signatureMatches(s *types.Signature, params typeCheckers, results typeCheckers) bool {
	return s != nil &&
		s.Params().Len() == len(params) &&
		s.Results().Len() == len(results) &&
		params.check(s.Params()) &&
		results.check(s.Results())
}

type typeCheckers []typeChecker

func (c typeCheckers) check(tuple *types.Tuple) bool {
	for i, checker := range c {
		if !checker(tuple.At(i).Type()) {
			return false
		}
	}
	return true
}

type typeChecker func(types.Type) bool

func getMethodSignature(pkg *types.Package, typ types.Type, name string) *types.Signature {
	ms := types.NewMethodSet(typ)
	method := ms.Lookup(pkg, name)
	if method == nil {
		return nil
	}

	return method.Obj().(*types.Func).Type().(*types.Signature)
}

func isEmptyInterface(typ types.Type) bool {
	switch typ := typ.(type) {
	case *types.Interface:
		return typ.NumMethods() == 0
	}
	return false
}

func isDriverValue(typ types.Type) bool {
	switch typ := typ.(type) {
	case *types.Named:
		return typ.String() == "database/sql/driver.Value"
	}
	return false
}

// isModel checks if the type is a model. If dive is true, it will check also
// the types of the struct if the type is a struct.
func isModel(typ types.Type, dive bool) bool {
	switch typ := typ.(type) {
	case *types.Named:
		if typeName(typ) == BaseModel {
			return true
		}
		return isModel(typ.Underlying(), true && dive)
	case *types.Pointer:
		return isModel(typ.Elem(), true && dive)
	case *types.Struct:
		if !dive {
			return false
		}

		for i := 0; i < typ.NumFields(); i++ {
			if isModel(typ.Field(i).Type(), false) {
				return true
			}
		}
	}
	return false
}

func (p *Processor) processBaseField(m *Model, f *Field) {
	m.Table = f.Tag.Get("table")
	if m.Table == "" {
		m.Table = toLowerSnakeCase(m.Name)
	}
}

func typeName(typ types.Type) string {
	return removeGoPath(typ.String())
}

var separator = filepath.Separator

// toSlash is an identical implementation of filepath.ToSlash. Is only
// implemented so we can change the separator on runtime for testing purposes,
// since filepath.Separator is a constant.
// Parts of the code using filepath.ToSlash that need cross-platform tests
// should use this instead.
func toSlash(path string) string {
	if separator == '/' {
		return path
	}
	return strings.Replace(path, string(separator), "/", -1)
}

func removeGoPath(path string) string {
	var prefix string
	if strings.HasPrefix(path, "[]*") {
		prefix = "[]*"
		path = path[3:]
	} else if strings.HasPrefix(path, "[]") {
		prefix = "[]"
		path = path[2:]
	} else if strings.HasPrefix(path, "*") {
		prefix = "*"
		path = path[1:]
	}

	path = toSlash(path)
	for _, p := range defaultGoPath {
		p = toSlash(p + "/src/")
		if strings.HasPrefix(path, p) {
			// Directories named "vendor" are only vendor directories
			// if they're under $GOPATH/src.
			if idx := strings.LastIndex(path, "/vendor/"); idx >= len(p)-1 {
				return prefix + path[idx+8:]
			}
			return prefix + path[len(p):]
		}
	}
	return prefix + path
}

func isIgnoredField(s *types.Struct, idx int) bool {
	tag := reflect.StructTag(s.Tag(idx))
	return strings.Split(tag.Get("kallax"), ",")[0] == "-"
}
