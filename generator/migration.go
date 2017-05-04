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
	Lock *DBSchema
}

// NewMigration creates a new migration from the old and the new schema.
func NewMigration(old, new *DBSchema) *Migration {
	var migration = &Migration{}
	migration.Up = SchemaDiff(old, new)
	migration.Down = migration.Up.ReverseChangeSet(old)
	migration.Lock = new
	return migration
}

// DBSchema represents a schema of all the models in the database.
type DBSchema struct {
	// Tables are the schema of all the tables.
	Tables []*TableSchema
}

// SchemaFromPackages returns a schema for the given packages models.
func SchemaFromPackages(pkgs ...*Package) (*DBSchema, error) {
	t := newPackageTransformer()
	return t.transform(pkgs...)
}

func (s *DBSchema) MarshalText() ([]byte, error) {
	schema := struct {
		Tables []*TableSchema
	}{s.Tables}
	return json.MarshalIndent(schema, "", "  ")
}

// Table finds a table with the given name.
func (s *DBSchema) Table(name string) *TableSchema {
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
		buf.WriteRune('\t')
		buf.WriteString(c.String())
		if i < len(s.Columns)-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteRune('\n')
		}
	}
	buf.WriteString(");\n\n")
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

func (s *TableSchema) Equals(s2 *TableSchema) bool {
	if s.Name != s2.Name || len(s.Columns) != len(s2.Columns) {
		return false
	}

	for i, c := range s.Columns {
		if !c.Equals(s2.Columns[i]) {
			return false
		}
	}

	return true
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

func (s *ColumnSchema) Equals(s2 *ColumnSchema) bool {
	return s.Name == s2.Name &&
		s.Type == s2.Type &&
		s.PrimaryKey == s2.PrimaryKey &&
		s.NotNull == s2.NotNull &&
		s.Reference.Equals(s2.Reference)
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

func NumericColumn(precision int) ColumnType {
	return ColumnType(fmt.Sprintf("numeric(%d)", precision))
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

func (r *Reference) Equals(r2 *Reference) bool {
	if r == nil && r2 == nil {
		return true
	} else if r == nil || r2 == nil {
		return false
	}

	return r.Table == r2.Table &&
		r.Column == r2.Column
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
		buf.WriteRune('\n')
	}
	return buf.Bytes(), nil
}

func (cs ChangeSet) String() string {
	var buf bytes.Buffer
	for _, c := range cs {
		buf.WriteString(fmt.Sprintf("- %s", c))
	}
	return buf.String()
}

// Reverse returns the change that will revert the current change set.
func (cs ChangeSet) Reverse(old *DBSchema) Change {
	var result = make(ChangeSet, len(cs))
	for i, c := range cs {
		result[i] = c.Reverse(old)
	}
	return result
}

// ReverseChangeSet returns the reverse change set of the current one.
func (cs ChangeSet) ReverseChangeSet(old *DBSchema) ChangeSet {
	return cs.Reverse(old).(ChangeSet)
}

// Change represents a change to be made in a migration.
type Change interface {
	encoding.TextMarshaler
	fmt.Stringer
	// Reverse returns the change that will revert the current change.
	Reverse(old *DBSchema) Change
}

// CreateTable is a change that will add a new table.
type CreateTable struct {
	*TableSchema
}

func (c *CreateTable) Reverse(old *DBSchema) Change {
	return &DropTable{Name: c.Name}
}

func (c *CreateTable) MarshalText() ([]byte, error) {
	return []byte(c.TableSchema.String()), nil
}

func (c *CreateTable) String() string {
	var cols = make([]string, len(c.Columns))
	for i, c := range c.Columns {
		cols[i] = c.Name
	}
	return fmt.Sprintf("A new table %q has been added with the following columns: %s.", c.Name, strings.Join(cols, ", "))
}

// DropTable is a change that will drop a table.
type DropTable struct {
	// Name is the name of the table to drop.
	Name string
}

func (c *DropTable) Reverse(old *DBSchema) Change {
	return &CreateTable{old.Table(c.Name)}
}

func (c *DropTable) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("DROP TABLE %s;\n", c.Name)), nil
}

func (c *DropTable) String() string {
	return fmt.Sprintf("Table %q has been deleted, and it will be dropped.", c.Name)
}

// AddColumn is a change that will add a column.
type AddColumn struct {
	// Column schema.
	Column *ColumnSchema
	// Table to add the column to.
	Table string
}

func (c *AddColumn) Reverse(old *DBSchema) Change {
	return &DropColumn{
		Table: c.Table,
		Name:  c.Column.Name,
	}
}

func (c *AddColumn) String() string {
	return fmt.Sprintf("A new column %q of type %q has been added to table %q.", c.Column.Name, c.Column.Type, c.Table)
}

func (c *AddColumn) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;\n", c.Table, c.Column)), nil
}

// DropColumn is a change that will drop a column.
type DropColumn struct {
	// Name of the column.
	Name string
	// Table name.
	Table string
}

func (c *DropColumn) Reverse(old *DBSchema) Change {
	return &AddColumn{
		Table:  c.Table,
		Column: old.Table(c.Table).Column(c.Name),
	}
}

func (c *DropColumn) String() string {
	return fmt.Sprintf("The column %q of table %q has been removed and it will be dropped.", c.Name, c.Table)
}

func (c *DropColumn) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", c.Table, c.Name)), nil
}

// ManualChange is a change that cannot be made automatically and requires
// the user to write a proper migration.
type ManualChange struct {
	Msg string
}

func (c *ManualChange) Reverse(old *DBSchema) Change {
	return c
}

func (c *ManualChange) String() string {
	return fmt.Sprintf("A manual change is required: %s.", c.Msg)
}

func (c *ManualChange) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("+++ THIS REQUIRES MANUAL MIGRATION: %s +++\n", c.Msg)), nil
}

// SchemaDiff generates a change set with the diff between two schemas.
func SchemaDiff(old, new *DBSchema) ChangeSet {
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

type packageTransformer struct {
	// pkg is the current package being transformed.
	pkg *Package
	// schema is the final schema being built.
	schema *DBSchema

	tables map[string]*TableSchema
	// tableIndex is a map from a Go type to a table name
	tableIndex map[string]string
	// pkIndex is a map from a table name to its primary key
	pkIndex map[string]*Field
	// inverses keeps all inverses indexed by type name
	// so they can be added later.
	inverses map[string][]*ColumnSchema
}

func newPackageTransformer() *packageTransformer {
	return &packageTransformer{
		schema:     new(DBSchema),
		tables:     make(map[string]*TableSchema),
		tableIndex: make(map[string]string),
		pkIndex:    make(map[string]*Field),
		inverses:   make(map[string][]*ColumnSchema),
	}
}

func (t *packageTransformer) transform(pkgs ...*Package) (*DBSchema, error) {
	for _, pkg := range pkgs {
		for _, m := range pkg.Models {
			t.tableIndex[m.Node.String()] = m.Table
			t.pkIndex[m.Table] = m.ID
		}
	}

	for _, pkg := range pkgs {
		t.pkg = pkg
		if err := t.transformPkg(pkg); err != nil {
			return nil, err
		}
	}

	if err := t.applyInverses(); err != nil {
		return nil, err
	}

	return t.schema, nil
}

func (t *packageTransformer) applyInverses() error {
	for typ, inverses := range t.inverses {
		table, ok := t.tableIndex[typ]
		if !ok {
			return fmt.Errorf("kallax: unable to find a table for model %s. Is the model package on the input for this command?", typ)
		}

		schema := t.tables[table]
		for _, inv := range inverses {
			if col := schema.Column(inv.Name); col != nil {
				if !col.Equals(inv) {
					return fmt.Errorf("kallax: there is an inverse definition conflicting with the column definition of column %s in the table %s. Please, make sure both definitions match.", inv.Name, table)
				}
			} else {
				schema.Columns = append(schema.Columns, inv)
			}
		}
	}

	return nil
}

func (t *packageTransformer) transformPkg(pkg *Package) error {
	for _, m := range pkg.Models {
		table, err := t.transformModel(m)
		if err != nil {
			return err
		}

		if prevTable, ok := t.tables[m.Table]; ok && !prevTable.Equals(table) {
			return fmt.Errorf("kallax: found more than one model for table %s", m.Table)
		}

		t.schema.Tables = append(t.schema.Tables, table)
		t.tables[table.Name] = table
	}
	return nil
}

func (t *packageTransformer) transformModel(m *Model) (*TableSchema, error) {
	schema := &TableSchema{Name: m.Table}
	var columns = make(map[string]*ColumnSchema)
	var err error
	schema.Columns, err = t.transformFields(m.Fields, columns)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (t *packageTransformer) transformFields(fields []*Field, columns map[string]*ColumnSchema) ([]*ColumnSchema, error) {
	var result []*ColumnSchema

	for _, f := range fields {
		if f.IsEmbedded {
			cols, err := t.transformFields(f.Fields, columns)
			if err != nil {
				return nil, err
			}
			result = append(result, cols...)
		} else {
			column, err := t.transformField(f)
			if err != nil {
				return nil, err
			}

			if f.Kind == Relationship && f.IsInverse() {
				typ := removeTypePrefix(f.Type)
				t.inverses[typ] = append(t.inverses[typ], column)
			} else if col, ok := columns[f.ColumnName()]; ok {
				if !col.Equals(column) {
					return nil, fmt.Errorf("kallax: there are two conflicting definitions for column %s on table %s: \n- %s\n- %s", col.Name, f.Model.Table, col, column)
				}
				// if it's the same column we can skip it
			} else {
				result = append(result, column)
				columns[column.Name] = column
			}
		}
	}

	return result, nil
}

func (t *packageTransformer) transformField(f *Field) (*ColumnSchema, error) {
	typ, err := t.transformType(f, f.IsPrimaryKey())
	if err != nil {
		return nil, err
	}

	ref, err := t.transformRef(f)
	if err != nil {
		return nil, err
	}

	name := f.ColumnName()
	if f.Kind == Relationship {
		name = f.ForeignKey()
	}

	return &ColumnSchema{
		Name:       name,
		PrimaryKey: f.IsPrimaryKey(),
		NotNull:    false,
		Type:       typ,
		Reference:  ref,
	}, nil
}

func (t *packageTransformer) transformType(f *Field, pk bool) (ColumnType, error) {
	if typ := f.SQLType(); typ != "" {
		return ColumnType(typ), nil
	}

	if f.IsJSON {
		return JSONBColumn, nil
	}

	if f.Kind == Array || f.Kind == Slice {
		return ArrayColumn(typeMappings[removeTypePrefix(f.Type)]), nil
	}

	if pk {
		if !isValidIdentifier(f) {
			return ColumnType(""), fmt.Errorf("kallax: type %s is not a valid type for a primary key. On field %s of model %s.", f.Type, f.Name, f.Model.Name)
		}

		return idTypeMappings[identifierType(f)], nil
	}

	if f.Kind == Basic {
		typ, ok := typeMappings[f.Type]
		if !ok {
			return ColumnType(""), fmt.Errorf("kallax: type %s can not be converted to a SQL type. On field %s of model %s. Consider using the struct tag `sqltype` to set a custom type for this column.", f.Type, f.Name, f.Model.Name)
		}
		return typ, nil
	}

	if f.Kind == Relationship && !f.IsInverse() {
		typ := removeTypePrefix(f.Type)
		table, ok := t.tableIndex[typ]
		if !ok {
			return ColumnType(""), fmt.Errorf("kallax: unable to find table for type %s in field %s of model %s. Is the model type part of the generation input?", typ, f.Name, f.Model.Name)
		}

		return t.transformType(t.pkIndex[table], false)
	}

	if f.Kind == Relationship {
		return t.transformType(f.Model.ID, false)
	}

	if f.Kind == Interface {
		typ := removeTypePrefix(typeName(f.Node.Type()))
		if typ, ok := typeMappings[typ]; ok {
			return typ, nil
		}
	}

	return ColumnType(""), fmt.Errorf("kallax: cannot find a suitable type (%s) for field %s of model %s. Consider using the struct tag `sqltype` to set a custom type for this column.", f.Type, f.Name, f.Model.Name)
}

func (t *packageTransformer) transformRef(f *Field) (*Reference, error) {
	if f.Kind == Relationship && !f.IsInverse() {
		typ := removeTypePrefix(f.Type)
		table, ok := t.tableIndex[typ]
		if !ok {
			return nil, fmt.Errorf("kallax: unable to find table for type %s in field %s of model %s. Is the model type part of the generation input?", typ, f.Name, f.Model.Name)
		}

		return &Reference{Table: table, Column: t.pkIndex[table].ColumnName()}, nil
	} else if f.Kind == Relationship {
		return &Reference{Table: f.Model.Table, Column: f.Model.ID.ColumnName()}, nil
	}

	return nil, nil
}

var typeMappings = map[string]ColumnType{
	"gopkg.in/src-d/go-kallax.v1.ULID":      UUIDColumn,
	"gopkg.in/src-d/go-kallax.v1.UUID":      UUIDColumn,
	"gopkg.in/src-d/go-kallax.v1.NumericID": BigIntColumn,
	"github.com/satori/go.uuid.UUID":        UUIDColumn,
	"string":        TextColumn,
	"rune":          ColumnType("char(1)"),
	"uint8":         SmallIntColumn,
	"int8":          SmallIntColumn,
	"byte":          SmallIntColumn,
	"uint16":        IntegerColumn,
	"int16":         SmallIntColumn,
	"uint32":        BigIntColumn,
	"int32":         IntegerColumn,
	"uint":          NumericColumn(20),
	"int":           BigIntColumn,
	"int64":         BigIntColumn,
	"uint64":        NumericColumn(20),
	"float32":       RealColumn,
	"float64":       DoubleColumn,
	"bool":          BooleanColumn,
	"url.URL":       TextColumn,
	"time.Time":     TimestamptzColumn,
	"time.Duration": BigIntColumn,
}

var idTypeMappings = map[string]ColumnType{
	"kallax.ULID":      UUIDColumn,
	"kallax.UUID":      UUIDColumn,
	"kallax.NumericID": SerialColumn,
}
