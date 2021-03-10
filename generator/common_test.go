package generator

import (
	"go/token"
	"go/types"
	"reflect"
	"testing"

	"golang.org/x/tools/go/packages"
)

func mkField(name, typ, tag string, fields ...*Field) *Field {
	f := NewField(name, typ, reflect.StructTag(tag))
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

func inline(f *Field) *Field {
	f.Tag = reflect.StructTag(`kallax:",inline"`)
	return f
}

func processorFixture(t *testing.T, source string) (*Processor, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports,
		Overlay: map[string][]byte{
			"fixture/fixture.go": []byte(source),
		},
	}, "github.com/loyalguru/go-kallax/generator/fixture")
	if err != nil {
		return nil, err
	}

	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		if len(pkg.Errors) > 0 {
			t.Fatalf("packages.Load had error in package %s: %v", pkg, pkg.Errors[0])
		}
	})

	p := pkgs[0].Types

	prc := NewProcessor("fixture", []string{"foo.go"})
	prc.Package = p
	return prc, nil
}

func processFixture(t *testing.T, source string) (*Package, error) {
	prc, err := processorFixture(t, source)
	if err != nil {
		return nil, err
	}

	prc.Silent()
	return prc.processPackage()
}
