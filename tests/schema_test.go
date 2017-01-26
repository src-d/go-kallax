package tests

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SchemaSuite struct {
	BaseTestSuite
}

func TestSchemaSuite(t *testing.T) {
	suite.Run(t, new(SchemaSuite))
}

func (s *SchemaSuite) TestSchemaID() {
	s.Equal("id", Schema.SchemaFixture.ID.String())
}

func (s *SchemaSuite) TestSchemaBasicField() {
	s.Equal("string", Schema.SchemaFixture.String.String())
}

func (s *SchemaSuite) TestSchemaRanamedField() {
	s.Equal("int", Schema.SchemaFixture.Int.String())
}

func (s *SchemaSuite) TestSchemaInlineField() {
	s.Equal("inline", Schema.SchemaFixture.Inline.String())
}

func (s *SchemaSuite) TestSchemaMapsOfString() {
	s.Equal("map_of_string", Schema.SchemaFixture.MapOfString.String())
}

func (s *SchemaSuite) TestSchemaMapOfSomeType() {
	s.Equal("map_of_interface", Schema.SchemaFixture.MapOfInterface.String())
}

func (s *SchemaSuite) TestSchemaMapOfInterface() {
	s.Equal("map_of_some_type", Schema.SchemaFixture.MapOfSomeType.String())
}

func (s *SchemaSuite) TestSchemaIgnored() {
	schema := reflect.ValueOf(Schema.SchemaFixture)
	field := reflect.Indirect(schema).FieldByName("ShouldIgnore")
	s.False(field.IsValid(), "TODO: https://github.com/src-d/go-kallax/issues/59")
}
