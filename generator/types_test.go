package generator

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
			"types.JSON(&r.Foo), nil",
		},
		{
			Map, true, false, "Foo", "", nil,
			"types.JSON(&r.Foo), nil",
		},
		{
			Struct, false, false, "Foo", "", nil,
			"&r.Foo, nil",
		},
		{
			Array, false, false, "Foo", "[5]string", nil,
			`types.Array(&r.Foo, 5), nil`,
		},
		{
			Slice, false, false, "Foo", "", nil,
			"types.Slice(&r.Foo), nil",
		},
		{
			Basic, false, true, "Foo", "", nil,
			"r.Foo, nil",
		},
		{
			Basic, false, true, "Foo", "", withParent(mkField("Bar", ""), mkField("Baz", "")),
			"r.Baz.Bar.Foo, nil",
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

func (s *ModelSuite) SetupSuite() {
	path := filepath.Join(os.Getenv("GOPATH"), "src", "github.com/src-d/go-kallax/fixtures")
	p := NewProcessor(path, nil)
	pkg, err := p.Do()
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

func TestModelValidate(t *testing.T) {
	require := require.New(t)
	m := &Model{Name: "Foo", Table: "foo"}
	m.Fields = []*Field{
		mkField("ID", ""),
		inline(mkField("Nested", "", inline(
			mkField("Deep", "", mkField("ID", "")),
		))),
	}
	require.NotNil(m.Validate(), "should return error")

	m.Fields = []*Field{
		mkField("ID", ""),
		inline(mkField("Nested", "", mkField("Foo", ""))),
	}
	require.Nil(m.Validate(), "should not return error")

	m.Table = ""
	require.NotNil(m.Validate(), "should return error")
}

func TestFieldForeignKey(t *testing.T) {
	assert := assert.New(t)
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

		assert.Equal(c.expected, f.ForeignKey(), "foreign key with tag: %s", c.tag)
		assert.Equal(c.inverse, f.IsInverse(), "is inverse: %s", c.tag)
	}
}

func TestModel(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}
