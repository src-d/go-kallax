package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestNewMigration(t *testing.T) {
	old := mkSchema(table1)
	new := mkSchema(table1, table2)
	migration, err := NewMigration(old, new)
	require.NoError(t, err)

	expectedUp := ChangeSet{&CreateTable{table2}}
	expectedDown := ChangeSet{&DropTable{"table2"}}

	require.Equal(t, expectedUp, migration.Up)
	require.Equal(t, expectedDown, migration.Down)
	require.Equal(t, migration.Lock, new)
}

func TestNewMigration_SelfRef(t *testing.T) {
	old := mkSchema()
	new := mkSchema(mkTable(
		"selfref",
		mkCol("id", SerialColumn, true, false, nil),
		mkCol("parent_id", BigIntColumn, false, false, mkRef("selfref", "id", true)),
		mkCol("child_id", BigIntColumn, false, false, mkRef("selfref", "id", false)),
	))

	_, err := NewMigration(old, new)
	require.NoError(t, err)
}

var table1 = mkTable(
	"table",
	mkCol("id", SerialColumn, true, true, nil),
	mkCol("num", DecimalColumn(1, 2), false, false, nil),
)

var table2 = mkTable(
	"table2",
	mkCol("table_id", SerialColumn, false, true, mkRef("table", "id", false)),
	mkCol("num", NumericColumn(20), false, false, nil),
)

const expectedTable = `CREATE TABLE table (
	id serial NOT NULL PRIMARY KEY,
	num decimal(1, 2)
);
`

const expectedTable2 = `CREATE TABLE table2 (
	table_id serial NOT NULL REFERENCES table(id),
	num numeric(20)
);
`

func TestTableSchema(t *testing.T) {
	require.Equal(t, expectedTable+"\n", table1.String())
	require.Equal(t, expectedTable2+"\n", table2.String())
}

func TestArrayColumn(t *testing.T) {
	require.Equal(t, ColumnType("text[]"), ArrayColumn(TextColumn))
	require.Equal(t, ColumnType("text[]"), ArrayColumn(ArrayColumn(TextColumn)))
}

func TestChangeSet(t *testing.T) {
	assertChange(
		t,
		ChangeSet{
			&DropTable{"foo"},
			&DropColumn{"col", "table"},
		},
		"BEGIN;\n\nDROP TABLE foo;\n\nALTER TABLE table DROP COLUMN col;\n\nCOMMIT;\n",
	)
}

func TestCreateTable(t *testing.T) {
	assertChange(
		t,
		&CreateTable{mkTable(
			"table",
			mkCol("foo", SmallIntColumn, false, false, nil),
			mkCol("bar", SerialColumn, false, false, nil),
		)},
		`CREATE TABLE table (
	foo smallint,
	bar serial
);

`)
}

func TestDropTable(t *testing.T) {
	assertChange(
		t,
		&DropTable{"table"},
		"DROP TABLE table;\n",
	)
}

func TestAddColumn(t *testing.T) {
	assertChange(
		t,
		&AddColumn{
			mkCol("foo", SmallIntColumn, false, false, nil),
			"table",
		},
		"ALTER TABLE table ADD COLUMN foo smallint;\n",
	)
}

func TestDropColumn(t *testing.T) {
	assertChange(
		t,
		&DropColumn{"col", "table"},
		"ALTER TABLE table DROP COLUMN col;\n",
	)
}

func TestManualChange(t *testing.T) {
	assertChange(
		t,
		&ManualChange{"foo"},
		"+++ THIS REQUIRES MANUAL MIGRATION: foo +++\n",
	)
}

func assertChange(t *testing.T, c Change, expected string) {
	output, err := c.MarshalText()
	require.NoError(t, err)
	require.Equal(t, expected, string(output))
}

func TestSchemaDiff(t *testing.T) {
	old := mkSchema(
		mkTable("removed"),
		mkTable(
			"shared",
			mkCol("foo", TextColumn, false, false, nil),
		),
	)

	new := mkSchema(
		mkTable(
			"shared",
			mkCol("foo", TextColumn, false, false, nil),
			mkCol("bar", TextColumn, false, false, nil),
		),
		mkTable("new"),
	)

	expected := ChangeSet{
		&DropTable{"removed"},
		&AddColumn{mkCol("bar", TextColumn, false, false, nil), "shared"},
		&CreateTable{mkTable("new")},
	}

	require.Equal(t, expected, SchemaDiff(old, new))
}

func TestTableSchemaDiff(t *testing.T) {
	old := mkTable(
		"table",
		mkCol("removed", TextColumn, false, false, nil),
		mkCol("shared", TextColumn, false, false, nil),
	)

	new := mkTable(
		"table",
		mkCol("new", TextColumn, false, false, nil),
		mkCol("shared", TextColumn, false, false, nil),
	)

	expected := ChangeSet{
		&DropColumn{"removed", "table"},
		&AddColumn{mkCol("new", TextColumn, false, false, nil), "table"},
	}

	require.Equal(t, expected, TableSchemaDiff(old, new))
}

func TestColumnSchemaDiff(t *testing.T) {
	cases := []struct {
		name                 string
		old, new             *ColumnSchema
		requiresManualChange bool
	}{
		{
			"type change",
			mkCol("foo", TextColumn, false, false, nil),
			mkCol("foo", SmallIntColumn, false, false, nil),
			true,
		},
		{
			"pk change",
			mkCol("foo", TextColumn, true, false, nil),
			mkCol("foo", TextColumn, false, false, nil),
			true,
		},
		{
			"not null change",
			mkCol("foo", TextColumn, false, true, nil),
			mkCol("foo", TextColumn, false, false, nil),
			true,
		},
		{
			"ref added",
			mkCol("foo", TextColumn, false, false, nil),
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar", false)),
			true,
		},
		{
			"ref removed",
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar", false)),
			mkCol("foo", TextColumn, false, false, nil),
			true,
		},
		{
			"ref table changed",
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar", false)),
			mkCol("foo", TextColumn, false, false, mkRef("bar", "bar", false)),
			true,
		},
		{
			"ref col changed",
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar", false)),
			mkCol("foo", TextColumn, false, false, mkRef("foo", "foo", false)),
			true,
		},
		{
			"ref col unchanged",
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar", false)),
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar", false)),
			false,
		},
		{
			"equal",
			mkCol("foo", TextColumn, false, false, nil),
			mkCol("foo", TextColumn, false, false, nil),
			false,
		},
	}

	for _, c := range cases {
		changes := ColumnSchemaDiff("Table", c.old, c.new)
		if !c.requiresManualChange {
			require.Len(t, changes, 0, c.name)
		} else {
			require.True(t, len(changes) > 0, c.name)
		}
	}
}

func TestReverseChange(t *testing.T) {
	require := require.New(t)
	old := mkSchema(
		mkTable(
			"foo",
			mkCol("bar", SmallIntColumn, false, false, nil),
		),
	)

	cases := []struct {
		original Change
		expected Change
	}{
		{
			&CreateTable{&TableSchema{Name: "foo"}},
			&DropTable{Name: "foo"},
		},
		{
			&DropTable{Name: "foo"},
			&CreateTable{old.Table("foo")},
		},
		{
			&AddColumn{
				Table:  "foo",
				Column: mkCol("bar", SmallIntColumn, false, false, nil),
			},
			&DropColumn{Table: "foo", Name: "bar"},
		},
		{
			&DropColumn{Table: "foo", Name: "bar"},
			&AddColumn{
				Table:  "foo",
				Column: mkCol("bar", SmallIntColumn, false, false, nil),
			},
		},
		{
			&ManualChange{"foo"},
			&ManualChange{"foo"},
		},
	}

	for _, c := range cases {
		require.Equal(c.expected, c.original.Reverse(old), "%T", c.original)
	}
}

func TestTableSchemaEquals(t *testing.T) {
	cases := []struct {
		name     string
		schema   *TableSchema
		expected bool
	}{
		{
			"different name",
			mkTable("bar"),
			false,
		},
		{
			"different number of columns",
			mkTable(
				"foo",
				mkCol("col1", IntegerColumn, false, false, nil),
				mkCol("col2", IntegerColumn, false, false, nil),
				mkCol("col3", IntegerColumn, false, false, nil),
			),
			false,
		},
		{
			"different column",
			mkTable(
				"foo",
				mkCol("col1", IntegerColumn, false, false, nil),
				mkCol("col4", IntegerColumn, false, false, nil),
			),
			false,
		},
		{
			"equal",
			mkTable(
				"foo",
				mkCol("col1", IntegerColumn, false, false, nil),
				mkCol("col2", IntegerColumn, false, false, nil),
			),
			true,
		},
	}

	schema := mkTable(
		"foo",
		mkCol("col1", IntegerColumn, false, false, nil),
		mkCol("col2", IntegerColumn, false, false, nil),
	)

	for _, c := range cases {
		require.Equal(t, c.expected, schema.Equals(c.schema), c.name)
	}
}

func TestColumnSchemaEquals(t *testing.T) {
	cases := []struct {
		name     string
		a, b     *ColumnSchema
		expected bool
	}{
		{
			"different name",
			mkCol("foo", TextColumn, false, false, nil),
			mkCol("bar", TextColumn, false, false, nil),
			false,
		},
		{
			"different pk",
			mkCol("id", SerialColumn, true, false, nil),
			mkCol("id", SerialColumn, false, false, nil),
			false,
		},
		{
			"different notnull",
			mkCol("foo", TextColumn, false, true, nil),
			mkCol("foo", TextColumn, false, false, nil),
			false,
		},
		{
			"one of the references is nil",
			mkCol("foo", TextColumn, false, false, nil),
			mkCol("foo", TextColumn, false, false, mkRef("a", "b", false)),
			false,
		},
		{
			"reference table does not match",
			mkCol("foo", TextColumn, false, false, mkRef("a", "b", false)),
			mkCol("foo", TextColumn, false, false, mkRef("b", "b", false)),
			false,
		},
		{
			"reference column does not match",
			mkCol("foo", TextColumn, false, false, mkRef("a", "b", false)),
			mkCol("foo", TextColumn, false, false, mkRef("a", "a", false)),
			false,
		},
		{
			"equal with reference",
			mkCol("foo", TextColumn, false, false, mkRef("a", "b", false)),
			mkCol("foo", TextColumn, false, false, mkRef("a", "b", false)),
			true,
		},
		{
			"equal without reference",
			mkCol("foo", TextColumn, false, false, nil),
			mkCol("foo", TextColumn, false, false, nil),
			true,
		},
	}

	for _, c := range cases {
		require.Equal(t, c.expected, c.a.Equals(c.b), c.name)
	}
}

func TestChangeSetSorted(t *testing.T) {
	old := mkSchema(
		mkTable("table2"),
		mkTable(
			"table1",
			mkCol("foo", SerialColumn, false, false, mkRef("table2", "bar", false)),
		),
		mkTable("table3"),
	)
	new := mkSchema(
		mkTable("table3"),
		mkTable("table4"),
		mkTable(
			"table5",
			mkCol("foo", SerialColumn, false, false, mkRef("table4", "bar", false)),
		),
	)
	cs := SchemaDiff(old, new)
	expected := ChangeSet{
		&CreateTable{new.Table("table4")},
		&CreateTable{new.Table("table5")},
		&DropTable{"table1"},
		&DropTable{"table2"},
	}

	sorted, err := cs.sorted(old.index(), new.index())
	require.NoError(t, err)
	require.Equal(t, expected, sorted)
}

type PackageTransformerSuite struct {
	suite.Suite
	t   *packageTransformer
	pkg *Package
}

const packageTransformerSourceFixture = `
package foo

import (
	"gopkg.in/src-d/go-kallax.v1"
	"net/url"
)

type User struct {
	kallax.Model ` + "`table:\"users\"`" + `
	ID kallax.ULID ` + "`pk:\"\"`" + `
	Username string
	// array field
	Emails []string
	Profile *Profile
}

type Profile struct {
	kallax.Model ` + "`table:\"profiles\"`" + `
	ID int64 ` + "`pk:\"autoincr\"`" + `
	Color string ` + "`sqltype:\"char(6)\"`" + `
	Background url.URL
	// should be added here because is an inverse and not
	// in user
	User *User ` + "`fk:\"user_id,inverse\"`" + `
	Spouse *kallax.ULID
	// a fk without reference in the other model
	// should be added anyway
	// should be added as bigint, as it is not a pk
	Metadata *ProfileMetadata
}

type ProfileMetadata struct {
	kallax.Model ` + "`table:\"metadata\"`" + `	
	// it's an pk, should be serial
	ID int64 ` + "`pk:\"autoincr\"`" + `
	// a json field
	Metadata map[string]interface{}
}
`

func (s *PackageTransformerSuite) SetupTest() {
	s.t = newPackageTransformer()
	var err error
	s.pkg, err = processFixture(packageTransformerSourceFixture)
	s.Require().NoError(err)
}

func (s *PackageTransformerSuite) TestMigrationCircularDep() {
	require := s.Require()
	schema, err := s.t.transform(s.pkg)
	require.NoError(err)
	require.NotNil(schema)

	_, err = NewMigration(mkSchema(), schema)
	require.NoError(err)
}

func (s *PackageTransformerSuite) TestTransform() {
	require := s.Require()
	schema, err := s.t.transform(s.pkg)
	require.NoError(err)
	require.NotNil(schema)

	expected := mkSchema(
		mkTable(
			"profiles",
			mkCol("id", SerialColumn, true, false, nil),
			mkCol("color", ColumnType("char(6)"), false, false, nil),
			mkCol("background", TextColumn, false, false, nil),
			mkCol("user_id", UUIDColumn, false, false, mkRef("users", "id", true)),
			mkCol("spouse", UUIDColumn, false, false, nil),
		),
		mkTable(
			"metadata",
			mkCol("id", SerialColumn, true, false, nil),
			mkCol("metadata", JSONBColumn, false, false, nil),
			mkCol("profile_id", BigIntColumn, false, false, mkRef("profiles", "id", false)),
		),
		mkTable(
			"users",
			mkCol("id", UUIDColumn, true, false, nil),
			mkCol("username", TextColumn, false, false, nil),
			mkCol("emails", ArrayColumn(TextColumn), false, false, nil),
		),
	)

	require.Equal(expected, schema)
}

func (s *PackageTransformerSuite) TestApplyInverses_TableNotFound() {
	s.t.fks["foo"] = []*ColumnSchema{
		mkCol("foo", TextColumn, false, false, nil),
	}

	s.Error(s.t.applyForeignKeys())
}

func (s *PackageTransformerSuite) TestApplyInverses_ConfictingCol() {
	s.t.fks["foo"] = []*ColumnSchema{
		mkCol("foo", TextColumn, false, false, nil),
	}
	s.t.tableIndex["foo"] = "foo"
	s.t.tables["foo"] = mkTable(
		"foo",
		mkCol("bar", TextColumn, false, false, nil),
		mkCol("foo", IntegerColumn, false, false, nil),
	)

	s.Error(s.t.applyForeignKeys())
}

func (s *PackageTransformerSuite) TestTransform_RepeatedTable() {
	m := *s.pkg.Models[len(s.pkg.Models)-1]
	m.Fields = nil
	s.pkg.Models = append(s.pkg.Models, &m)

	_, err := s.t.transform(s.pkg)
	s.Error(err)
}

func TestPackageTransformer(t *testing.T) {
	suite.Run(t, new(PackageTransformerSuite))
}

func TestGraphResolve(t *testing.T) {
	require := require.New(t)
	g := newGraph().
		add("a").
		add("b").
		dependsOn("c", "b").
		dependsOn("d", "c").
		dependsOn("e", "b").
		dependsOn("e", "a")

	result, err := g.resolve()
	require.NoError(err)
	require.Equal([]string{"e", "a", "d", "c", "b"}, result)

	g = newGraph().
		add("a").
		dependsOn("c", "b").
		dependsOn("a", "c").
		dependsOn("b", "a")

	_, err = g.resolve()
	require.Error(err)
}

func mkSchema(tables ...*TableSchema) *DBSchema {
	return &DBSchema{tables}
}

func mkTable(name string, columns ...*ColumnSchema) *TableSchema {
	return &TableSchema{name, columns}
}

func mkCol(name string, typ ColumnType, pk, notNull bool, ref *Reference) *ColumnSchema {
	return &ColumnSchema{name, typ, pk, ref, notNull, false}
}

func mkColUnique(name string, typ ColumnType, pk, notNull bool, ref *Reference) *ColumnSchema {
	return &ColumnSchema{name, typ, pk, ref, notNull, true}
}

func mkRef(table, col string, inverse bool) *Reference {
	return &Reference{table, col, inverse}
}
