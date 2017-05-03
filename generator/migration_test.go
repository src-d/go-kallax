package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMigration(t *testing.T) {
	old := mkModel(table1)
	new := mkModel(table1, table2)
	migration := NewMigration(old, new)

	expectedUp := ChangeSet{&CreateTable{table2}}
	expectedDown := ChangeSet{&DropTable{"table2"}}

	require.Equal(t, expectedUp, migration.Up)
	require.Equal(t, expectedDown, migration.Down)
	require.Equal(t, migration.Lock, new)
}

var table1 = mkTable(
	"table",
	mkCol("id", SerialColumn, true, true, nil),
	mkCol("num", DecimalColumn(1, 2), false, false, nil),
)

var table2 = mkTable(
	"table2",
	mkCol("table_id", SerialColumn, false, true, mkRef("table", "id")),
	mkCol("num", NumericColumn(1, 2), false, false, nil),
)

const expectedTable = `CREATE TABLE table (
id serial NOT NULL PRIMARY KEY,
num decimal(1, 2)
);
`

const expectedTable2 = `CREATE TABLE table2 (
table_id serial NOT NULL REFERENCES table(id),
num numeric(1, 2)
);
`

func TestTableSchema(t *testing.T) {

	require.Equal(t, expectedTable, table1.String())
	require.Equal(t, expectedTable2, table2.String())
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
		"DROP TABLE foo;\nALTER TABLE table DROP COLUMN col;\n",
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
	old := mkModel(
		mkTable("removed"),
		mkTable(
			"shared",
			mkCol("foo", TextColumn, false, false, nil),
		),
	)

	new := mkModel(
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
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar")),
			true,
		},
		{
			"ref removed",
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar")),
			mkCol("foo", TextColumn, false, false, nil),
			true,
		},
		{
			"ref table changed",
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar")),
			mkCol("foo", TextColumn, false, false, mkRef("bar", "bar")),
			true,
		},
		{
			"ref col changed",
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar")),
			mkCol("foo", TextColumn, false, false, mkRef("foo", "foo")),
			true,
		},
		{
			"ref col unchanged",
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar")),
			mkCol("foo", TextColumn, false, false, mkRef("foo", "bar")),
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
	old := mkModel(
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
		require.Equal(c.expected, reverseChange(c.original, old), "%T", c.original)
	}
}

func mkModel(tables ...*TableSchema) *ModelSchema {
	return &ModelSchema{tables}
}

func mkTable(name string, columns ...*ColumnSchema) *TableSchema {
	return &TableSchema{name, columns}
}

func mkCol(name string, typ ColumnType, pk, notNull bool, ref *Reference) *ColumnSchema {
	return &ColumnSchema{name, typ, pk, ref, notNull}
}

func mkRef(table, col string) *Reference {
	return &Reference{table, col}
}
