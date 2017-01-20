package kallax

import "fmt"

// Schema represents a table schema in the database. Contains some information
// like the table name, its columns, its identifier and so on.
type Schema interface {
	// Alias returns the name of the alias used in queries for this schema.
	Alias() string
	// Table returns the table name.
	Table() string
	// ID returns the name of the identifier of the table.
	ID() SchemaField
	// Columns returns the list of columns in the schema.
	Columns() []SchemaField
}

// BaseSchema is the basic implementation of Schema.
type BaseSchema struct {
	alias   string
	table   string
	id      SchemaField
	columns []SchemaField
}

// NewBaseSchema creates a new schema with the given table, alias, identifier
// and columns.
func NewBaseSchema(table, alias string, id SchemaField, columns ...SchemaField) *BaseSchema {
	return &BaseSchema{
		alias:   alias,
		table:   table,
		id:      id,
		columns: columns,
	}
}

func (s *BaseSchema) Alias() string          { return s.alias }
func (s *BaseSchema) Table() string          { return s.table }
func (s *BaseSchema) ID() SchemaField        { return s.id }
func (s *BaseSchema) Columns() []SchemaField { return s.columns }

// SchemaField is a named field in the table schema.
type SchemaField interface {
	isSchemaField()
	// String returns the string representation of the field. That is, its name.
	String() string
	// QualifiedString returns the name of the field qualified by the alias of
	// the given schema.
	QualifiedName(Schema) string
}

// BaseSchemaField is a basic schema field with name.
type BaseSchemaField struct {
	name string
}

// NewSchemaField creates a new schema field with the given name.
func NewSchemaField(name string) SchemaField {
	return &BaseSchemaField{name}
}

func (*BaseSchemaField) isSchemaField() {}

func (f BaseSchemaField) String() string {
	return f.name
}

func (f *BaseSchemaField) QualifiedName(schema Schema) string {
	alias := schema.Alias()
	if alias != "" {
		return fmt.Sprintf("%s.%s", alias, f.name)
	}
	return f.name
}

// ColumnNames returns the names of the given schema fields.
func ColumnNames(columns []SchemaField) []string {
	var names = make([]string, len(columns))
	for i, v := range columns {
		names[i] = v.String()
	}
	return names
}
