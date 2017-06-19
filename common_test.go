package kallax

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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

var tableSchemas = []string{
	`CREATE TABLE IF NOT EXISTS model (
		id serial PRIMARY KEY,
		name varchar(255) not null,
		email varchar(255) not null,
		age int not null
	)`,
	`CREATE TABLE IF NOT EXISTS rel (
		id serial PRIMARY KEY,
		model_id integer,
		foo text
	)`,
	`CREATE TABLE IF NOT EXISTS through_left (
		id serial PRIMARY KEY,
		name text not null
	)`,
	`CREATE TABLE IF NOT EXISTS through_right (
		id serial PRIMARY KEY,
		name text not null
	)`,
	`CREATE TABLE IF NOT EXISTS through_middle (
		id serial PRIMARY KEY,
		left_id bigint references through_left(id),
		right_id bigint references through_right(id)
	)`,
}

func setupTables(t *testing.T, db *sql.DB) {
	for _, ts := range tableSchemas {
		_, err := db.Exec(ts)
		require.NoError(t, err)
	}
}

var tableNames = []string{
	"rel", "model", "through_middle", "through_left", "through_right",
}

func teardownTables(t *testing.T, db *sql.DB) {
	for _, tn := range tableNames {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE %s", tn))
		require.NoError(t, err)
	}
}

type model struct {
	Model
	ID    int64 `pk:"autoincr"`
	Name  string
	Email string
	Age   int
	Rel   *rel
	Rels  []*rel
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
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
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
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *model) NewRelationshipRecord(field string) (Record, error) {
	switch field {
	case "rel":
		return new(rel), nil
	case "rels":
		return new(rel), nil
	}
	return nil, fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *model) SetRelationship(field string, record interface{}) error {
	switch field {
	case "rel":
		rel, ok := record.(*rel)
		if !ok {
			return fmt.Errorf("kallax: can't set relationship %s with a record of type %t", field, record)
		}
		m.Rel = rel
		return nil
	case "rels":
		rels, ok := record.([]Record)
		if !ok {
			return fmt.Errorf("kallax: can't set relationship %s with value of type %T", field, record)
		}
		m.Rels = make([]*rel, len(rels))
		for i, r := range rels {
			rel, ok := r.(*rel)
			if !ok {
				return fmt.Errorf("kallax: can't set element of relationship %s with element of type %T", field, r)
			}
			m.Rels[i] = rel
		}
		return nil
	}
	return fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *model) GetID() Identifier {
	return (*NumericID)(&m.ID)
}

type rel struct {
	Model
	ID  int64 `pk:"autoincr"`
	Foo string
}

func newRel(id Identifier, foo string) *rel {
	rel := &rel{NewModel(), 0, foo}
	rel.AddVirtualColumn("model_id", id)
	return rel
}

func (r *rel) GetID() Identifier {
	return (*NumericID)(&r.ID)
}

func (m *rel) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return m.ID, nil
	case "model_id":
		return m.VirtualColumn(col), nil
	case "foo":
		return m.Foo, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *rel) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return &m.ID, nil
	case "model_id":
		return VirtualColumn(col, m, new(NumericID)), nil
	case "foo":
		return &m.Foo, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *rel) NewRelationshipRecord(field string) (Record, error) {
	return nil, fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *rel) SetRelationship(field string, record interface{}) error {
	return fmt.Errorf("kallax: no relationship found for field %s", field)
}

type onlyPkModel struct {
	Model
	ID int64 `pk:"autoincr"`
}

func newOnlyPkModel() *onlyPkModel {
	m := new(onlyPkModel)
	return m
}

func (m *onlyPkModel) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return m.ID, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *onlyPkModel) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return &m.ID, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *onlyPkModel) NewRelationshipRecord(field string) (Record, error) {
	return nil, fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *onlyPkModel) SetRelationship(field string, record interface{}) error {
	return fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *onlyPkModel) GetID() Identifier {
	return (*NumericID)(&m.ID)
}

var ModelSchema = NewBaseSchema(
	"model",
	"__model",
	f("id"),
	ForeignKeys{
		"rel":     []*ForeignKey{NewForeignKey("model_id", false)},
		"rels":    []*ForeignKey{NewForeignKey("model_id", false)},
		"rel_inv": []*ForeignKey{NewForeignKey("model_id", true)},
	},
	func() Record {
		return new(model)
	},
	true,
	f("id"),
	f("name"),
	f("email"),
	f("age"),
)

var RelSchema = NewBaseSchema(
	"rel",
	"__rel",
	f("id"),
	ForeignKeys{},
	func() Record {
		return new(rel)
	},
	true,
	f("id"),
	f("model_id"),
	f("foo"),
)

var onlyPkModelSchema = NewBaseSchema(
	"model",
	"__model",
	f("id"),
	nil,
	func() Record {
		return new(onlyPkModel)
	},
	true,
	f("id"),
)

type throughLeft struct {
	Model
	ID     int64
	Name   string
	Rights []*throughRight
}

func newThroughLeft(name string) *throughLeft {
	m := &throughLeft{Model: NewModel(), Name: name}
	return m
}

func (m *throughLeft) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return m.ID, nil
	case "name":
		return m.Name, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *throughLeft) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return &m.ID, nil
	case "name":
		return &m.Name, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *throughLeft) NewRelationshipRecord(field string) (Record, error) {
	switch field {
	case "Rights":
		return new(throughRight), nil
	}
	return nil, fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *throughLeft) SetRelationship(field string, record interface{}) error {
	switch field {
	case "Rights":
		rels, ok := record.([]Record)
		if !ok {
			return fmt.Errorf("kallax: can't set relationship %s with value of type %T", field, record)
		}
		m.Rights = make([]*throughRight, len(rels))
		for i, r := range rels {
			rel, ok := r.(*throughRight)
			if !ok {
				return fmt.Errorf("kallax: can't set element of relationship %s with element of type %T", field, r)
			}
			m.Rights[i] = rel
		}
		return nil
	}
	return fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *throughLeft) GetID() Identifier {
	return (*NumericID)(&m.ID)
}

type throughMiddle struct {
	Model
	ID    int64
	Left  *throughLeft
	Right *throughRight
}

func newThroughMiddle(left *throughLeft, right *throughRight) *throughMiddle {
	m := &throughMiddle{Model: NewModel(), Left: left, Right: right}
	return m
}

func (m *throughMiddle) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return m.ID, nil
	case "left_id":
		return m.Left.ID, nil
	case "right_id":
		return m.Right.ID, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *throughMiddle) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return &m.ID, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *throughMiddle) NewRelationshipRecord(field string) (Record, error) {
	switch field {
	case "Left":
		return new(throughLeft), nil
	case "Right":
		return new(throughRight), nil
	}
	return nil, fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *throughMiddle) SetRelationship(field string, record interface{}) error {
	switch field {
	case "Left":
		rel, ok := record.(*throughLeft)
		if !ok {
			return fmt.Errorf("kallax: can't set relationship %s with a record of type %t", field, record)
		}
		m.Left = rel
		return nil
	case "Right":
		rel, ok := record.(*throughRight)
		if !ok {
			return fmt.Errorf("kallax: can't set relationship %s with a record of type %t", field, record)
		}
		m.Right = rel
		return nil
	}
	return fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *throughMiddle) GetID() Identifier {
	return (*NumericID)(&m.ID)
}

type throughRight struct {
	Model
	ID    int64
	Name  string
	Lefts []*throughLeft
}

func newThroughRight(name string) *throughRight {
	m := &throughRight{Model: NewModel(), Name: name}
	return m
}

func (m *throughRight) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return m.ID, nil
	case "name":
		return m.Name, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *throughRight) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return &m.ID, nil
	case "name":
		return &m.Name, nil
	}
	return nil, fmt.Errorf("kallax: column does not exist: %s", col)
}

func (m *throughRight) NewRelationshipRecord(field string) (Record, error) {
	switch field {
	case "Lefts":
		return new(throughLeft), nil
	}
	return nil, fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *throughRight) SetRelationship(field string, record interface{}) error {
	switch field {
	case "Lefts":
		rels, ok := record.([]Record)
		if !ok {
			return fmt.Errorf("kallax: can't set relationship %s with value of type %T", field, record)
		}
		m.Lefts = make([]*throughLeft, len(rels))
		for i, r := range rels {
			rel, ok := r.(*throughLeft)
			if !ok {
				return fmt.Errorf("kallax: can't set element of relationship %s with element of type %T", field, r)
			}
			m.Lefts[i] = rel
		}
		return nil
	}
	return fmt.Errorf("kallax: no relationship found for field %s", field)
}

func (m *throughRight) GetID() Identifier {
	return (*NumericID)(&m.ID)
}

var ThroughLeftSchema = NewBaseSchema(
	"through_left",
	"__thleft",
	f("id"),
	ForeignKeys{
		"Rights": []*ForeignKey{
			NewForeignKey("left_id", false),
			NewForeignKey("right_id", false),
		},
	},
	func() Record {
		return new(throughLeft)
	},
	true,
	f("id"),
	f("name"),
)

var ThroughMiddleSchema = NewBaseSchema(
	"through_middle",
	"__thmiddle",
	f("id"),
	ForeignKeys{
		"Left":  []*ForeignKey{NewForeignKey("left_id", false)},
		"Right": []*ForeignKey{NewForeignKey("right_id", false)},
	},
	func() Record {
		return new(throughMiddle)
	},
	true,
	f("id"),
	f("left_id"),
	f("right_id"),
)

var ThroughRightSchema = NewBaseSchema(
	"through_right",
	"__thright",
	f("id"),
	ForeignKeys{
		"Lefts": []*ForeignKey{
			NewForeignKey("right_id", false),
			NewForeignKey("left_id", false),
		},
	},
	func() Record {
		return new(throughRight)
	},
	true,
	f("id"),
	f("name"),
)

func f(name string) SchemaField {
	return NewSchemaField(name)
}
