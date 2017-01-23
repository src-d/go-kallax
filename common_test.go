package kallax

import (
	"database/sql"
	"fmt"
	"os"
)

func envOrDefault(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		v = def
	}
	return v
}

func openTestDB() (*sql.DB, error) {
	return sql.Open("postgres", fmt.Sprintf(
		"postgres://%s:%s@0.0.0.0:5432/%s?sslmode=disable",
		envOrDefault("DBUSER", "testing"),
		envOrDefault("DBPASS", "testing"),
		envOrDefault("DBNAME", "testing"),
	))
}

type model struct {
	Model
	Name  string
	Email string
	Age   int
	Rel   *rel
}

func newModel(name, email string, age int) *model {
	m := &model{Model: NewModel(), Name: name, Email: email, Age: age}
	return m
}

func (m *model) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return m.ID, nil
	case "name":
		return m.Name, nil
	case "email":
		return m.Email, nil
	case "age":
		return m.Age, nil
	}
	return nil, fmt.Errorf("column does not exist: %s", col)
}

func (m *model) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return &m.ID, nil
	case "name":
		return &m.Name, nil
	case "email":
		return &m.Email, nil
	case "age":
		return &m.Age, nil
	}
	return nil, fmt.Errorf("column does not exist: %s", col)
}

func (m *model) NewRelationshipRecord(field string) (Record, error) {
	switch field {
	case "rel":
		return new(rel), nil
	}
	return nil, fmt.Errorf("no relationship found for field %s", field)
}

func (m *model) SetRelationship(field string, record Record) error {
	switch field {
	case "rel":
		rel, ok := record.(*rel)
		if !ok {
			return fmt.Errorf("can't set relationship %s with a record of type %t", field, record)
		}
		m.Rel = rel
		return nil
	}
	return fmt.Errorf("no relationship found for field %s", field)
}

type rel struct {
	Model
	ModelID ID
	Foo     string
}

func newRel(id ID, foo string) *rel {
	return &rel{NewModel(), id, foo}
}

func (m *rel) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return m.ID, nil
	case "model_id":
		return m.ModelID, nil
	case "foo":
		return m.Foo, nil
	}
	return nil, fmt.Errorf("column does not exist: %s", col)
}

func (m *rel) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return &m.ID, nil
	case "model_id":
		return &m.ModelID, nil
	case "foo":
		return &m.Foo, nil
	}
	return nil, fmt.Errorf("column does not exist: %s", col)
}

func (m *rel) NewRelationshipRecord(field string) (Record, error) {
	return nil, fmt.Errorf("no relationship found for field %s", field)
}

func (m *rel) SetRelationship(field string, record Record) error {
	return fmt.Errorf("no relationship found for field %s", field)
}

var ModelSchema = &BaseSchema{
	alias: "__model",
	table: "model",
	id:    f("id"),
	foreignKeys: ForeignKeys{
		"rel": NewSchemaField("model_id"),
	},
	columns: []SchemaField{
		f("id"),
		f("name"),
		f("email"),
		f("age"),
	},
}

var RelSchema = &BaseSchema{
	alias:       "__rel",
	table:       "rel",
	id:          f("id"),
	foreignKeys: ForeignKeys{},
	columns: []SchemaField{
		f("id"),
		f("model_id"),
		f("foo"),
	},
}

func f(name string) SchemaField {
	return NewSchemaField(name)
}
