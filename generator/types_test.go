package generator

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/tools/go/packages"
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
		s.Equal(c.inline, mkField("", c.typ, c.tag).Inline(), "field with tag: %s", c.tag)
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
		s.Equal(c.ok, mkField("", "", c.tag).IsPrimaryKey(), "field with tag: %s", c.tag)
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
		s.Equal(c.ok, mkField("", "", c.tag).IsAutoIncrement(), "field with tag: %s", c.tag)
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
		{`kallax:"foo"`, "Bar", "foo"},
		{`kallax:"References"`, "Bar", "_References"},
		{`kallax:"references"`, "Bar", "_references"},
	}

	for _, c := range cases {
		name := mkField(c.name, "", c.tag).ColumnName()
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
			Slice, false, false, "Foo", "[]string", nil,
			"types.Slice(&r.Foo)",
		},
		{
			Basic, false, true, "Foo", "", nil,
			"types.Nullable(&r.Foo)",
		},
		{
			Interface, false, true, "Foo", "", nil,
			"types.Nullable(r.Foo)",
		},
		{
			Basic, false, true, "Foo", "", withParent(mkField("Bar", "", ""), mkField("Baz", "", "")),
			"types.Nullable(&r.Baz.Bar.Foo)",
		},
	}

	for i, c := range cases {
		f := withKind(withParent(mkField(c.name, c.typeStr, ""), c.parent), c.kind)
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
			mkField("Foo", "string", ""),
			"r.Foo, nil",
		},
		{
			withAlias(mkField("Foo", "string", "")),
			"(string)(r.Foo), nil",
		},
		{
			withPtr(withAlias(mkField("Foo", "string", ""))),
			"(*string)(r.Foo), nil",
		},
		{
			withKind(mkField("Foo", "", ""), Slice),
			"types.Slice(r.Foo), nil",
		},
		{
			withKind(mkField("Foo", "[5]string", ""), Array),
			`types.Array(&r.Foo, 5), nil`,
		},
		{
			withJSON(withKind(mkField("Foo", "", ""), Map)),
			"types.JSON(r.Foo), nil",
		},
		{
			withKind(mkField("Foo", "", ""), Struct),
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
package fixture

import (
	"errors"
	"strings"

	kallax "github.com/networkteam/go-kallax"
)

type User struct {
	kallax.Model ` + "`table:\"users\" pk:\"id\"`" + `
	ID kallax.ULID
	Username     string
	Email        string
	Password     Password
	Websites     []string
	Emails       []*Email
	Settings     *Settings
}

func newUser(id kallax.ULID, username, email string, websites []string) (*User, error) {
	if strings.Contains(email, "@spam.org") {
		return nil, errors.New("kallax: is spam!")
	}
	return &User{ID: id, Username: username, Email: email, Websites: websites}, nil
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
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports,
		Overlay: map[string][]byte{
			"fixture/fixture.go": []byte(fixturesSource),
		},
	}, "github.com/networkteam/go-kallax/generator/fixture")
	s.NoError(err)

	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		if len(pkg.Errors) > 0 {
			s.NoError(pkg.Errors[0], "packages.Load had errors in package %s", pkg)
		}
	})

	p := pkgs[0].Types

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
	s.Equal("id kallax.ULID, username string, email string, websites []string", s.model.CtorArgs())
	s.Equal("id, username, email, websites", s.model.CtorArgVars())
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
		mkField("ID", "", ""),
		inline(mkField("Nested", "", "", inline(
			mkField("Deep", "", "", mkField("ID", "", "")),
		))),
	}
	require.Error(m.Validate(), "should return error")

	m.Fields = []*Field{
		mkField("ID", "", ""),
		inline(mkField("Nested", "", "", mkField("Foo", "", ""))),
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

func (s *ModelSuite) TestString() {
	s.Equal("\"Variadic\" [3 Field(s)] [Events: []]", s.variadic.String())
	s.Equal("\"User\" [7 Field(s)] [Events: []]", s.model.String())
}

func TestFieldForeignKey(t *testing.T) {
	r := require.New(t)
	m := &Model{Name: "Foo", Table: "bar", Type: "foo.Foo"}

	cases := []struct {
		tag      string
		inverse  bool
		typ      string
		expected string
	}{
		{`fk:""`, false, "", "foo_id"},
		{`fk:"foo_bar_baz"`, false, "", "foo_bar_baz"},
		{`fk:",inverse"`, true, "Bar", "bar_id"},
		{`fk:"foos,inverse"`, true, "Bar", "foos"},
		{``, false, "", "foo_id"},
	}

	for _, c := range cases {
		f := NewField("", "", reflect.StructTag(c.tag))
		f.Kind = Relationship
		f.Model = m
		f.Type = c.typ

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
		id     string
	}{
		{
			"only one primary key",
			[]*Field{
				mkField("Foo", "", ""),
				mkField("ID", "", `pk:""`),
			},
			false,
			"ID",
		},
		{
			"multiple primary keys",
			[]*Field{
				mkField("ID", "", `pk:""`),
				mkField("FooID", "", `pk:""`),
			},
			true,
			"",
		},
		{
			"primary key defined in model but empty",
			[]*Field{
				mkField("Model", BaseModel, `pk:""`),
			},
			true,
			"",
		},
		{
			"primary key defined in model and non existent",
			[]*Field{
				mkField("Model", BaseModel, `pk:"foo"`),
				mkField("Bar", "", ""),
			},
			true,
			"",
		},
		{
			"primary key defined in model",
			[]*Field{
				mkField("Model", BaseModel, `pk:"foo"`),
				mkField("Baz", "", ""),
				mkField("Foo", "", ""),
				mkField("Bar", "", ""),
			},
			false,
			"Foo",
		},
	}

	for _, c := range cases {
		m := new(Model)
		err := m.SetFields(c.fields)
		if c.err {
			r.Error(err, c.name)
		} else {
			r.NoError(err, c.name)
			r.Equal(c.id, m.ID.Name)
		}
	}
}

func TestModel(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

func TestPkProperties(t *testing.T) {
	cases := []struct {
		tag          string
		name         string
		autoincr     bool
		isPrimaryKey bool
	}{
		{`pk:"bar"`, "bar", false, true},
		{`pk:""`, "", false, true},
		{`pk:"autoincr"`, "", true, true},
		{`pk:",autoincr"`, "", true, true},
		{`bar:"baz" pk:"foo"`, "foo", false, true},
		{`pk:"foo,autoincr"`, "foo", true, true},
	}

	require := require.New(t)
	for _, tt := range cases {
		name, autoincr, isPrimaryKey := pkProperties(reflect.StructTag(tt.tag))
		require.Equal(tt.name, name, tt.tag)
		require.Equal(tt.autoincr, autoincr, tt.tag)
		require.Equal(tt.isPrimaryKey, isPrimaryKey, tt.tag)
	}
}

func TestIsUnique(t *testing.T) {
	cases := []struct {
		tag    string
		unique bool
	}{
		{``, false},
		{`fk:"foo"`, false},
		{`unique:""`, false},
		{`unique:"true"`, true},
		{`fk:"foo" unique:"true"`, true},
	}

	for _, tt := range cases {
		t.Run(tt.tag, func(t *testing.T) {
			f := NewField("", "", reflect.StructTag(tt.tag))
			require.Equal(t, tt.unique, f.IsUnique())
		})
	}
}
