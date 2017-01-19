package generator

import (
	"os"
	"path/filepath"
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
		parent   *Field
		expected string
	}{
		{
			Struct, true, false, "Foo", nil,
			"types.JSON(&r.Foo), nil",
		},
		{
			Map, true, false, "Foo", nil,
			"types.JSON(&r.Foo), nil",
		},
		{
			Struct, false, false, "Foo", nil,
			"&r.Foo, nil",
		},
		{
			Array, false, false, "Foo", nil,
			`nil, fmt.Errorf("array types are not supported")`,
		},
		{
			Slice, false, false, "Foo", nil,
			"types.Array(&r.Foo), nil",
		},
		{
			Basic, false, true, "Foo", nil,
			"r.Foo, nil",
		},
		{
			Basic, false, true, "Foo", withParent(mkField("Bar", ""), mkField("Baz", "")),
			"r.Baz.Bar.Foo, nil",
		},
	}

	for i, c := range cases {
		f := withKind(withParent(mkField(c.name, ""), c.parent), c.kind)
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
			"types.Array(r.Foo), nil",
		},
		{
			withKind(mkField("Foo", ""), Array),
			`nil, fmt.Errorf("array go type not supported")`,
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
	model *Model
}

func (s *ModelSuite) SetupSuite() {
	path := filepath.Join(os.Getenv("GOPATH"), "src", "github.com/src-d/go-kallax/fixtures")
	p := NewProcessor(path, nil)
	pkg, err := p.Do()
	s.Nil(err)

	s.Len(pkg.Models, 2, "there should exist 2 models")
	for _, m := range pkg.Models {
		if m.Name == "User" {
			s.model = m
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

func TestModelValidateEvents(t *testing.T) {
	require := require.New(t)
	cases := []struct {
		events Events
		err    bool
	}{
		{Events{BeforeSave, AfterSave}, false},
		{Events{BeforeSave, BeforeInsert}, true},
		{Events{BeforeSave, BeforeUpdate}, true},
		{Events{AfterSave, AfterInsert}, true},
		{Events{AfterSave, AfterUpdate}, true},
	}

	for _, c := range cases {
		m := &Model{Table: "foo", Events: c.events}

		err := m.Validate()
		if c.err {
			require.NotNil(err, "%v", c.events)
			require.Equal(ErrEventConflict, err, "%v", c.events)
		} else {
			require.Nil(err, "%v", c.events)
		}
	}
}

func TestModel(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}
