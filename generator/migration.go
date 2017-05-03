package generator

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"strings"
)

// Migration contains all the data to represent a schema migration.
type Migration struct {
	// Up contains the changes to update from the previous version to the current one.
	Up ChangeSet
	// Down contains all the changes to downgrade to the previous version.
	Down ChangeSet
	// Lock contains the locked model schema.
	Lock *ModelSchema
}

// NewMigration creates a new migration from the old and the new schema.
func NewMigration(old, new *ModelSchema) *Migration {
	var migration = &Migration{}
	migration.Up = SchemaDiff(old, new)
	migration.Down = ReverseChangeSet(migration.Up, old)
	migration.Lock = new
	return migration
}

// ModelSchema represents a schema of all the models in the database.
type ModelSchema struct {
	// Tables are the schema of all the tables.
	Tables []*TableSchema
}

func (s *ModelSchema) MarshalText() ([]byte, error) {
	schema := struct {
		Tables []*TableSchema
	}{s.Tables}
	return json.MarshalIndent(schema, "", "  ")
}

// Table finds a table with the given name.
func (s *ModelSchema) Table(name string) *TableSchema {
	for _, t := range s.Tables {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// TableSchema represents the SQL schema of a table.
type TableSchema struct {
	// Name is the table name.
	Name string
	// Columns are the schemas of the columns in the table.
	Columns []*ColumnSchema
}

func (s *TableSchema) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", s.Name))
	for i, c := range s.Columns {
		buf.WriteString(c.String())
		if i < len(s.Columns)-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteRune('\n')
		}
	}
	buf.WriteString(");\n")
	return buf.String()
}

// Columns returns the schema of the column with the given name.
func (s *TableSchema) Column(name string) *ColumnSchema {
	for _, c := range s.Columns {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// ColumnSchema represents the schema of a column.
type ColumnSchema struct {
	// Name of the column.
	Name string
	// Type of the column.
	Type ColumnType
	// PrimaryKey reports whether the column is a primary key.
	PrimaryKey bool
	// Reference is an optional reference to another table column.
	// If it's not nil, it means this column has a foreign key.
	Reference *Reference
	// NotNull reports whether the column is not nullable.
	NotNull bool
}

func (s *ColumnSchema) String() string {
	var buf bytes.Buffer
	buf.WriteString(s.Name)
	buf.WriteRune(' ')
	buf.WriteString(string(s.Type))

	if s.NotNull {
		buf.WriteString(" NOT NULL")
	}

	if s.PrimaryKey {
		buf.WriteString(" PRIMARY KEY")
	}

	if s.Reference != nil {
		buf.WriteString(" REFERENCES ")
		buf.WriteString(s.Reference.String())
	}

	return buf.String()
}

// ColumnType represents the SQL column type.
type ColumnType string

const (
	SmallIntColumn    ColumnType = "smallint"
	IntegerColumn     ColumnType = "integer"
	BigIntColumn      ColumnType = "bigint"
	RealColumn        ColumnType = "real"
	DoubleColumn      ColumnType = "double"
	SmallSerialColumn ColumnType = "smallserial"
	SerialColumn      ColumnType = "serial"
	BigSerialColumn   ColumnType = "bigserial"
	TimestamptzColumn ColumnType = "timestamptz"
	TextColumn        ColumnType = "text"
	JSONBColumn       ColumnType = "jsonb"
	BooleanColumn     ColumnType = "boolean"
	UUIDColumn        ColumnType = "uuid"
)

func NumericColumn(precision, scale int) ColumnType {
	return ColumnType(fmt.Sprintf("numeric(%d, %d)", precision, scale))
}

func DecimalColumn(precision, scale int) ColumnType {
	return ColumnType(fmt.Sprintf("decimal(%d, %d)", precision, scale))
}

func ArrayColumn(typ ColumnType) ColumnType {
	// only allow arrays, not matrixes
	if strings.HasSuffix(string(typ), "[]") {
		return typ
	}
	return typ + "[]"
}

// Reference represents a reference to another table column.
type Reference struct {
	// Table is the referenced table.
	Table string
	// Column is the referenced column.
	Column string
}

func (r *Reference) String() string {
	return fmt.Sprintf("%s(%s)", r.Table, r.Column)
}

// ChangeSet is a set of changes to be made in a migration.
type ChangeSet []Change

func (cs ChangeSet) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	for _, c := range cs {
		bytes, err := c.MarshalText()
		if err != nil {
			return nil, err
		}
		buf.Write(bytes)
	}
	return buf.Bytes(), nil
}

// Change represents a change to be made in a migration.
type Change interface {
	encoding.TextMarshaler
}

// CreateTable is a change that will add a new table.
type CreateTable struct {
	*TableSchema
}

func (c *CreateTable) MarshalText() ([]byte, error) {
	return []byte(c.TableSchema.String()), nil
}

// DropTable is a change that will drop a table.
type DropTable struct {
	// Name is the name of the table to drop.
	Name string
}

func (c *DropTable) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("DROP TABLE %s;\n", c.Name)), nil
}

// AddColumn is a change that will add a column.
type AddColumn struct {
	// Column schema.
	Column *ColumnSchema
	// Table to add the column to.
	Table string
}

func (c *AddColumn) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN ", c.Table))
	buf.WriteString(c.Column.String())
	buf.WriteString(";\n")
	return buf.Bytes(), nil
}

// DropColumn is a change that will drop a column.
type DropColumn struct {
	// Name of the column.
	Name string
	// Table name.
	Table string
}

func (c *DropColumn) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", c.Table, c.Name)), nil
}

// ManualChange is a change that cannot be made automatically and requires
// the user to write a proper migration.
type ManualChange struct {
	Msg string
}

func (c *ManualChange) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("+++ THIS REQUIRES MANUAL MIGRATION: %s +++\n", c.Msg)), nil
}

// SchemaDiff generates a change set with the diff between two schemas.
func SchemaDiff(old, new *ModelSchema) ChangeSet {
	var cs ChangeSet
	for _, oldTable := range old.Tables {
		if t := new.Table(oldTable.Name); t == nil {
			cs = append(cs, &DropTable{Name: oldTable.Name})
		} else {
			cs = append(cs, TableSchemaDiff(oldTable, t)...)
		}
	}

	for _, newTable := range new.Tables {
		if t := old.Table(newTable.Name); t == nil {
			cs = append(cs, &CreateTable{newTable})
		}
	}

	return cs
}

// TableSchemaDiff generates a change set with the diff between two table
// schemas.
func TableSchemaDiff(old, new *TableSchema) ChangeSet {
	var cs ChangeSet
	for _, oldCol := range old.Columns {
		if c := new.Column(oldCol.Name); c == nil {
			cs = append(cs, &DropColumn{
				Table: old.Name,
				Name:  oldCol.Name,
			})
		} else {
			cs = append(cs, ColumnSchemaDiff(old.Name, oldCol, c)...)
		}
	}

	for _, newCol := range new.Columns {
		if c := old.Column(newCol.Name); c == nil {
			cs = append(cs, &AddColumn{
				Table:  new.Name,
				Column: newCol,
			})
		}
	}
	return cs
}

// ColumnSchemaDiff generates the change set with the diff between two column
// schemas.
func ColumnSchemaDiff(table string, old, new *ColumnSchema) ChangeSet {
	var cs ChangeSet
	if old.Type != new.Type {
		cs = append(cs, &ManualChange{
			fmt.Sprintf("don't know how to generate migration for a change of type in %s(%s)", table, new.Name),
		})
	}

	if old.PrimaryKey != new.PrimaryKey {
		cs = append(cs, &ManualChange{
			fmt.Sprintf("don't know how to generate migration for a change of primary key in %s(%s)", table, new.Name),
		})
	}

	if old.NotNull != new.NotNull {
		cs = append(cs, &ManualChange{
			fmt.Sprintf("don't know how to generate migration for a change of null/not null in %s(%s)", table, new.Name),
		})
	}

	if referenceChanged(old, new) {
		cs = append(cs, &ManualChange{
			fmt.Sprintf("don't know how to generate migration for a change of foreign key in %s(%s)", table, new.Name),
		})
	}

	return cs
}

func referenceChanged(old, new *ColumnSchema) bool {
	return old.Reference != new.Reference &&
		(old.Reference == nil ||
			new.Reference == nil ||
			old.Reference.Column != new.Reference.Column ||
			old.Reference.Table != new.Reference.Table)
}

// ReverseChangeSet returns a new change set with the inverses of the given
// change set.
func ReverseChangeSet(cs ChangeSet, old *ModelSchema) ChangeSet {
	var result = make(ChangeSet, len(cs))
	for i, c := range cs {
		result[i] = reverseChange(c, old)
	}
	return result
}

// reverseChange generates the inverse of a change.
func reverseChange(c Change, old *ModelSchema) Change {
	switch c := c.(type) {
	case *CreateTable:
		return &DropTable{Name: c.Name}

	case *DropTable:
		return &CreateTable{old.Table(c.Name)}

	case *AddColumn:
		return &DropColumn{
			Table: c.Table,
			Name:  c.Column.Name,
		}

	case *DropColumn:
		return &AddColumn{
			Table:  c.Table,
			Column: old.Table(c.Table).Column(c.Name),
		}

	case *ManualChange:
		return &ManualChange{
			Msg: c.Msg,
		}
	}

	panic("unreachable")
}
