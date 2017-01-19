package kallax

import "fmt"

// Schema represents a table schema in the database. Contains some information
// like the table name, its columns, its identifier and so on.
type Schema interface {
	// GetAlias returns the name of the alias used in queries for this schema.
	GetAlias() string
	// GetTable returns the table name.
	GetTable() string
	// GetID returns the name of the identifier of the table.
	GetID() SchemaField
	// GetColumns returns the list of columns in the schema.
	GetColumns() []SchemaField
}

// BaseSchema is the
type BaseSchema struct {
	Alias   string
	Table   string
	ID      SchemaField
	Columns []SchemaField
}

func (s *BaseSchema) GetAlias() string          { return s.Alias }
func (s *BaseSchema) GetTable() string          { return s.Table }
func (s *BaseSchema) GetID() SchemaField        { return s.ID }
func (s *BaseSchema) GetColumns() []SchemaField { return s.Columns }

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
	alias := schema.GetAlias()
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
