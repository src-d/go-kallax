package tests

import (
	"database/sql"
	"reflect"

	"github.com/stretchr/testify/suite"
)

type SchemaSuite struct {
	suite.Suite
	db *sql.DB
}

func (s *SchemaSuite) TestSchemaBasicField() {
	s.Equal("string", Schema.SchemaFixture.String)
}

func (s *SchemaSuite) TestSchemaRanamedField() {
	s.Equal("int", Schema.SchemaFixture.Int)
}

func (s *SchemaSuite) TestSchemaInlineField() {
	schema := reflect.ValueOf(Schema.SchemaFixture)
	field := reflect.Indirect(schema).FieldByName("Inline")
	s.True(field.IsValid())
}

func (s *SchemaSuite) TestSchemaMapsOfString() {
	schema := reflect.ValueOf(Schema.SchemaFixture)
	field := reflect.Indirect(schema).FieldByName("MapOfString")
	s.True(field.IsValid())
}

func (s *SchemaSuite) TestSchemaMapOfSomeType() {
	schema := reflect.ValueOf(Schema.SchemaFixture)
	field := reflect.Indirect(schema).FieldByName("MapOfSomeType")
	s.True(field.IsValid())
}

func (s *SchemaSuite) TestSchemaMapOfInterface() {
	schema := reflect.ValueOf(Schema.SchemaFixture)
	field := reflect.Indirect(schema).FieldByName("MapOfInterface")
	s.True(field.IsValid())
}
