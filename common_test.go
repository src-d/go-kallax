package kallax

import "database/sql"

func openTestDB() (*sql.DB, error) {
	return sql.Open("postgres", "postgres://testing:testing@0.0.0.0:5432/testing?sslmode=disable")
}

var ModelSchema = new(modelSchema)

type modelSchema struct{}

func (*modelSchema) Alias() string   { return "__model" }
func (*modelSchema) Table() string   { return "model" }
func (*modelSchema) ID() SchemaField { return f("id") }
func (*modelSchema) Columns() []SchemaField {
	return []SchemaField{
		f("id"),
		f("name"),
		f("email"),
		f("age"),
	}
}

func f(name string) SchemaField {
	return NewSchemaField(name)
}
