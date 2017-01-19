package generator

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/src-d/proteus/source"
)

// BaseModel is the type name of the kallax base model.
const BaseModel = "github.com/src-d/go-kallax.Model"

// Processor is in charge of processing the package in a patch and
// scan models from it.
type Processor struct {
	// Path of the package.
	Path string
	// Ignore is the list of files to ignore when scanning.
	Ignore map[string]struct{}
	// Package is the scanned package.
	Package *types.Package
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

// Do performs all the processing and returns the scanned package.
func (p *Processor) Do() (*Package, error) {
	files, err := p.getSourceFiles()
	if err != nil {
		return nil, err
	}

	p.Package, err = p.parseSourceFiles(files)
	if err != nil {
		return nil, err
	}

	return p.processPackage()
}

func (p *Processor) getSourceFiles() ([]string, error) {
	pkg, err := build.Default.ImportDir(p.Path, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot process directory %s: %s", p.Path, err)
	}

	var files []string
	files = append(files, pkg.GoFiles...)
	files = append(files, pkg.CgoFiles...)

	if len(files) == 0 {
		return nil, fmt.Errorf("%s: no buildable Go files", p.Path)
	}

	return joinDirectory(p.Path, p.removeIgnoredFiles(files)), nil
}

func (p *Processor) removeIgnoredFiles(filenames []string) []string {
	var output []string
	for _, filename := range filenames {
		if _, ok := p.Ignore[filename]; ok {
			continue
		}

		output = append(output, filename)
	}

	return output
}

func (p *Processor) parseSourceFiles(filenames []string) (*types.Package, error) {
	var files []*ast.File
	fs := token.NewFileSet()
	for _, filename := range filenames {
		file, err := parser.ParseFile(fs, filename, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parsing package: %s: %s", filename, err)
		}

		files = append(files, file)
	}

	config := types.Config{
		FakeImportC: true,
		Error:       func(error) {},
		Importer:    source.NewImporter(),
	}

	return config.Check(p.Path, fs, files, new(types.Info))
}

func (p *Processor) processPackage() (*Package, error) {
	pkg := &Package{Name: p.Package.Name()}
	var ctors []*types.Func

	fmt.Println("Package: ", pkg.Name)

	s := p.Package.Scope()
	for _, name := range s.Names() {
		obj := s.Lookup(name)
		switch t := obj.Type().(type) {
		case *types.Signature:
			if strings.HasPrefix(name, "new") {
				ctors = append(ctors, obj.(*types.Func))
			}
		case *types.Named:
			if str, ok := t.Underlying().(*types.Struct); ok {
				if m := p.processModel(name, str, t); m != nil {
					fmt.Printf("Found: %s\n", m)
					if err := m.Validate(); err != nil {
						return nil, err
					}

					pkg.Models = append(pkg.Models, m)
					m.Node = t
					m.Package = p.Package
				}
			}
		}
	}

	for _, ctor := range ctors {
		p.tryMatchConstructor(pkg.Models, ctor)
	}

	return pkg, nil
}

func (p *Processor) tryMatchConstructor(models []*Model, fun *types.Func) {
	for _, m := range models {
		if fun.Name() != fmt.Sprintf("new%s", m.Name) {
			continue
		}

		sig := fun.Type().(*types.Signature)
		if sig.Recv() != nil {
			continue
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
		return
	}
}

func (p *Processor) processModel(name string, s *types.Struct, t *types.Named) *Model {
	m := NewModel(name)
	m.Events = p.findEvents(t)

	var base int
	if base, m.Fields = p.processFields(s, nil, true); base == -1 {
		return nil
	}

	p.processBaseField(m, m.Fields[base])
	return m
}

func (p *Processor) findEvents(node *types.Named) []Event {
	var events []Event
	all := []Event{
		BeforeInsert, AfterInsert, BeforeUpdate, AfterUpdate, BeforeSave, AfterSave,
	}

	for _, e := range all {
		if p.isEventPresent(node, e) {
			events = append(events, e)
		}
	}

	return events
}

// isEventPresent checks the given Event is implemented for the given node.
// All Before events are required to have an error return type. All After
// events are required to have no return types.
func (p *Processor) isEventPresent(node *types.Named, e Event) bool {
	ms := types.NewMethodSet(types.NewPointer(node))
	method := ms.Lookup(p.Package, string(e))
	if method != nil {
		fn := method.Obj().(*types.Func)
		signature := fn.Type().(*types.Signature)
		if signature.Params().Len() > 0 {
			return false
		}

		if strings.HasPrefix(string(e), "Before") {
			if signature.Results().Len() != 1 ||
				!isBuiltinError(signature.Results().At(0).Type()) {
				return false
			}
			return true
		}

		if signature.Results().Len() > 0 {
			return false
		}

		return true
	}

	return false
}

// processFields returns which field index is an embedded kallax.Model, or -1 if none.
func (p *Processor) processFields(s *types.Struct, done []*types.Struct, root bool) (base int, fields []*Field) {
	base = -1

	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		if !f.Exported() {
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

		p.processField(field, f.Type(), done, root)
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
		} else {
			if t, ok := specialTypes[typeName(typ)]; ok {
				field.Type = t
				return
			}

			p.processField(field, typ.Underlying(), done, root)
			field.IsAlias = !field.IsJSON
		}
	case *types.Array:
		var underlying Field
		p.processField(&underlying, typ.Elem(), done, root)
		if underlying.Kind == Relationship {
			field.Kind = Relationship
			return
		}

		if underlying.Kind != Basic {
			field.IsJSON = true
		}
		field.Kind = Array
		field.Fields = underlying.Fields
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
		field.Fields = underlying.Fields
	case *types.Map:
		field.Kind = Map
		field.IsJSON = true
	case *types.Interface:
		field.Kind = Interface
		field.IsJSON = true
		// TODO: Check if implements SQLType, and then don't set IsJSON to true
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
		fmt.Printf("Ignored field %s of type %s\n", field.Name, field.Type)
	}
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

func joinDirectory(directory string, files []string) []string {
	result := make([]string, len(files))
	for i, file := range files {
		result[i] = filepath.Join(directory, file)
	}

	return result
}

var goPath = os.Getenv("GOPATH")

func typeName(typ types.Type) string {
	return strings.Replace(typ.String(), goPath+"/src/", "", -1)
}
