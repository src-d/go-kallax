package generator

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FieldSuite struct {
	suite.Suite
}

func TestField(t *testing.T) {
	suite.Run(t, new(FieldSuite))
}

func (s *FieldSuite) TestInline() {
	cases := []struct {
		typ    string
		tag    string
		inline bool
	}{
		{"", "", false},
		{BaseModel, "", true},
		{"", `kallax:"foo"`, false},
		{"", `kallax:"foo,inline"`, true},
		{"", `kallax:"foo,inline,omitempty"`, true},
		{"", `kallax:",inline,omitempty"`, true},
		{"", `kallax:",inline"`, true},
	}

	for _, c := range cases {
		s.Equal(c.inline, withTag(mkField("", c.typ), c.tag).Inline(), "field with tag: %s", c.tag)
	}
}

func (s *FieldSuite) TestIsPrimaryKey() {
	cases := []struct {
		tag string
		ok  bool
	}{
		{"", false},
		{`kallax:"pk"`, false},
		{`kallax:"foo,pk"`, false},
		{`pk:""`, true},
		{`pk:"foo"`, true},
		{`pk:"autoincr"`, true},
	}

	for _, c := range cases {
		s.Equal(c.ok, withTag(mkField("", ""), c.tag).IsPrimaryKey(), "field with tag: %s", c.tag)
	}
}

func (s *FieldSuite) TestIsAutoIncrement() {
	cases := []struct {
		tag string
		ok  bool
	}{
		{"", false},
		{`pk:""`, false},
		{`pk:"ponies"`, false},
		{`pk:"autoincr"`, true},
	}

	for _, c := range cases {
		s.Equal(c.ok, withTag(mkField("", ""), c.tag).IsAutoIncrement(), "field with tag: %s", c.tag)
	}
}

func (s *FieldSuite) TestColumnName() {
	cases := []struct {
		tag      string
		name     string
		expected string
	}{
		{"", "Foo", "foo"},
		{"", "FooBar", "foo_bar"},
		{"", "ID", "id"},
		{"", "References", "_references"},
		{`column:"foo"`, "Bar", "foo"},
		{`column:"References"`, "Bar", "_References"},
		{`column:"references"`, "Bar", "_references"},
	}

	for _, c := range cases {
		name := withTag(mkField(c.name, ""), c.tag).ColumnName()
		s.Equal(c.expected, name, "field with name: %q and tag: %s", c.name, c.tag)
	}
}

func (s *FieldSuite) TestAddress() {
	cases := []struct {
		kind     FieldKind
		isJSON   bool
		isPtr    bool
		name     string
		typeStr  string
		parent   *Field
		expected string
	}{
		{
			Struct, true, false, "Foo", "", nil,
			"types.JSON(&r.Foo)",
		},
		{
			Map, true, false, "Foo", "", nil,
			"types.JSON(&r.Foo)",
		},
		{
			Struct, false, false, "Foo", "", nil,
			"&r.Foo",
		},
		{
			Array, false, false, "Foo", "[5]string", nil,
			`types.Array(&r.Foo, 5)`,
		},
		{
			Slice, false, false, "Foo", "", nil,
			"types.Slice(&r.Foo)",
		},
		{
			Basic, false, true, "Foo", "", nil,
			"types.Nullable(&r.Foo)",
		},
		{
			Basic, false, true, "Foo", "", withParent(mkField("Bar", ""), mkField("Baz", "")),
			"types.Nullable(&r.Baz.Bar.Foo)",
		},
	}

	for i, c := range cases {
		f := withKind(withParent(mkField(c.name, c.typeStr), c.parent), c.kind)
		if c.isJSON {
			f = withJSON(f)
		}

		if c.isPtr {
			f = withPtr(f)
		}

		s.Equal(c.expected, f.Address(), "Field %s, i = %d", f.Name, i)
	}
}

func (s *FieldSuite) TestValue() {
	cases := []struct {
		field    *Field
		expected string
	}{
		{
			mkField("Foo", "string"),
			"r.Foo, nil",
		},
		{
			withAlias(mkField("Foo", "string")),
			"(string)(r.Foo), nil",
		},
		{
			withPtr(withAlias(mkField("Foo", "string"))),
			"(*string)(r.Foo), nil",
		},
		{
			withKind(mkField("Foo", ""), Slice),
			"types.Slice(r.Foo), nil",
		},
		{
			withKind(mkField("Foo", "[5]string"), Array),
			`types.Array(&r.Foo, 5), nil`,
		},
		{
			withJSON(withKind(mkField("Foo", ""), Map)),
			"types.JSON(r.Foo), nil",
		},
		{
			withKind(mkField("Foo", ""), Struct),
			"r.Foo, nil",
		},
	}

	for i, c := range cases {
		s.Equal(c.expected, c.field.Value(), "Field %s, i=%d", c.field.Name, i)
	}
}

type ModelSuite struct {
	suite.Suite
	model    *Model
	variadic *Model
}

const fixturesSource = `
package fixtures

import (
	"errors"
	"strings"

	kallax "github.com/src-d/go-kallax"
)

type User struct {
	kallax.Model ` + "`table:\"users\"`" + `
	ID int64 ` + "`pk:\"autoincr\"`" + `
	Username     string
	Email        string
	Password     Password
	Websites     []string
	Emails       []*Email
	Settings     *Settings
}

func newUser(username, email string) (*User, error) {
	if strings.Contains(email, "@spam.org") {
		return nil, errors.New("kallax: is spam!")
	}
	return &User{Username: username, Email: email}, nil
}

type Email struct {
	kallax.Model ` + "`table:\"models\"`" + `
	ID int64 ` + "`pk:\"autoincr\"`" + `
	Address      string
	Primary      bool
}

func newProfile(address string, primary bool) *Email {
	return &Email{Address: address, Primary: primary}
}

type Password string

// Kids, don't do this at home
func (p *Password) Set(pwd string) {
	*p = Password("such cypher" + pwd + "much secure")
}

type Settings struct {
	NotificationsActive bool
	NotifyByEmail       bool
}

type Variadic struct {
	kallax.Model
	ID int64 ` + "`pk:\"autoincr\"`" + `
	Foo []string
	Bar string
}

func newVariadic(bar string, foo ...string) *Variadic {
	return &Variadic{Foo: foo, Bar: bar}
}
`

func (s *ModelSuite) SetupSuite() {
	fset := &token.FileSet{}
	astFile, err := parser.ParseFile(fset, "fixture.go", fixturesSource, 0)
	s.Nil(err)

	cfg := &types.Config{
		Importer: importer.For("gc", nil),
	}
	p, err := cfg.Check("foo", fset, []*ast.File{astFile}, nil)
	s.Nil(err)

	prc := NewProcessor("fixture", []string{"foo.go"})
	prc.Package = p
	pkg, err := prc.processPackage()
	s.Nil(err)

	s.Len(pkg.Models, 3, "there should exist 3 models")
	for _, m := range pkg.Models {
		if m.Name == "User" {
			s.model = m
		}

		if m.Name == "Variadic" {
			s.variadic = m
		}
	}
	s.NotNil(s.model, "User struct should be defined")
}

func (s *ModelSuite) TestModel() {
	s.Equal("__user", s.model.Alias())
	s.Equal("users", s.model.Table)
	s.Equal("User", s.model.Name)
	s.Equal("UserStore", s.model.StoreName)
	s.Equal("UserQuery", s.model.QueryName)
	s.Equal("UserResultSet", s.model.ResultSetName)
}

func (s *ModelSuite) TestCtor() {
	s.Equal("username string, email string", s.model.CtorArgs())
	s.Equal("username, email", s.model.CtorArgVars())
	s.Equal("(record *User, err error)", s.model.CtorReturns())
	s.Equal("record, err", s.model.CtorRetVars())
}

func (s *ModelSuite) TestCtor_Variadic() {
	s.Equal("bar string, foo ...string", s.variadic.CtorArgs())
	s.Equal("bar, foo...", s.variadic.CtorArgVars())
	s.Equal("(record *Variadic)", s.variadic.CtorReturns())
	s.Equal("record", s.variadic.CtorRetVars())
}

func (s *ModelSuite) TestModelValidate() {
	require := s.Require()

	id := s.model.ID
	m := &Model{Name: "Foo", Table: "foo", ID: id}
	m.Fields = []*Field{
		mkField("ID", ""),
		inline(mkField("Nested", "", inline(
			mkField("Deep", "", mkField("ID", "")),
		))),
	}
	require.Error(m.Validate(), "should return error")

	m.Fields = []*Field{
		mkField("ID", ""),
		inline(mkField("Nested", "", mkField("Foo", ""))),
	}
	require.NoError(m.Validate(), "should not return error")

	m.ID = nil
	require.Error(m.Validate(), "should return error")

	m.ID = s.model.Fields[2]
	require.Error(m.Validate(), "should return error")

	m.ID = id
	m.Table = ""
	require.Error(m.Validate(), "should return error")
}

func TestFieldForeignKey(t *testing.T) {
	r := require.New(t)
	m := &Model{Name: "Foo", Table: "bar", Type: "foo.Foo"}

	cases := []struct {
		tag      string
		inverse  bool
		expected string
	}{
		{`fk:""`, false, "foo_id"},
		{`fk:"foo_bar_baz"`, false, "foo_bar_baz"},
		{`fk:",inverse"`, true, "foo_id"},
		{`fk:"foos,inverse"`, true, "foos"},
		{``, false, "foo_id"},
	}

	for _, c := range cases {
		f := NewField("", "", reflect.StructTag(c.tag))
		f.Kind = Relationship
		f.Model = m

		r.Equal(c.expected, f.ForeignKey(), "foreign key with tag: %s", c.tag)
		r.Equal(c.inverse, f.IsInverse(), "is inverse: %s", c.tag)
	}
}

func TestModelSetFields(t *testing.T) {
	r := require.New(t)
	cases := []struct {
		name   string
		fields []*Field
		err    bool
	}{
		{
			"only one primary key",
			[]*Field{
				mkField("Foo", ""),
				withTag(mkField("ID", ""), `pk:""`),
			},
			false,
		},
		{
			"multiple primary keys",
			[]*Field{
				withTag(mkField("ID", ""), `pk:""`),
				withTag(mkField("FooID", ""), `pk:""`),
			},
			true,
		},
	}

	for _, c := range cases {
		m := new(Model)
		err := m.SetFields(c.fields)
		if c.err {
			r.Error(err, c.name)
		} else {
			r.NoError(err, c.name)
		}
	}
}

func TestModel(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}
