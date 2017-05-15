package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"

	parseutil "gopkg.in/src-d/go-parse-utils.v1"
)

func mkField(name, typ string, fields ...*Field) *Field {
	f := NewField(name, typ, reflect.StructTag(""))
	f.SetFields(fields)
	return f
}

func withKind(f *Field, kind FieldKind) *Field {
	f.Kind = kind
	return f
}

func withPtr(f *Field) *Field {
	f.IsPtr = true
	return f
}

func withAlias(f *Field) *Field {
	f.IsAlias = true
	return f
}

func withJSON(f *Field) *Field {
	f.IsJSON = true
	return f
}

func withParent(f *Field, parent *Field) *Field {
	f.Parent = parent
	return f
}

func withNode(f *Field, name string, typ types.Type) *Field {
	f.Node = types.NewVar(token.NoPos, nil, name, typ)
	return f
}

func withTag(f *Field, tag string) *Field {
	f.Tag = reflect.StructTag(tag)
	return f
}

func inline(f *Field) *Field {
	return withTag(f, `kallax:",inline"`)
}

func processorFixture(source string) (*Processor, error) {
	fset := &token.FileSet{}
	astFile, err := parser.ParseFile(fset, "fixture.go", source, 0)
	if err != nil {
		return nil, err
	}

	cfg := &types.Config{
		Importer: parseutil.NewImporter(),
	}
	p, err := cfg.Check("foo", fset, []*ast.File{astFile}, nil)
	if err != nil {
		return nil, err
	}

	prc := NewProcessor("fixture", []string{"foo.go"})
	prc.Package = p
	return prc, nil
}

func processFixture(source string) (*Package, error) {
	prc, err := processorFixture(source)
	if err != nil {
		return nil, err
	}

	prc.Silent()
	return prc.processPackage()
}
