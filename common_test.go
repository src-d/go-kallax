package kallax

import "database/sql"

func openTestDB() (*sql.DB, error) {
	return sql.Open("postgres", "postgres://testing:testing@0.0.0.0:5432/testing?sslmode=disable")
}

var ModelSchema = new(modelSchema)

type modelSchema struct{}

func (*modelSchema) GetAlias() string   { return "__model" }
func (*modelSchema) GetTable() string   { return "model" }
func (*modelSchema) GetID() SchemaField { return f("id") }
func (*modelSchema) GetColumns() []SchemaField {
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
