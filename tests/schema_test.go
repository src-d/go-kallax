package tests

import "reflect"

func (s *CommonSuite) TestSchemaBasicField() {
	s.Equal(Schema.SchemaFixture.String, "string")
}

func (s *CommonSuite) TestSchemaRanamedField() {
	s.Equal(Schema.SchemaFixture.Int, "int")
}

func (s *CommonSuite) TestSchemaInlineField() {
	schema := reflect.ValueOf(Schema.SchemaFixture)
	field := reflect.Indirect(schema).FieldByName("Inline")
	s.True(field.IsValid())
}

func (s *CommonSuite) TestSchemaMapsOfString() {
	schema := reflect.ValueOf(Schema.SchemaFixture)
	field := reflect.Indirect(schema).FieldByName("MapOfString")
	s.True(field.IsValid())
}

func (s *CommonSuite) TestSchemaMapOfSomeType() {
	schema := reflect.ValueOf(Schema.SchemaFixture)
	field := reflect.Indirect(schema).FieldByName("MapOfSomeType")
	s.True(field.IsValid())
}

func (s *CommonSuite) TestSchemaMapOfInterface() {
	schema := reflect.ValueOf(Schema.SchemaFixture)
	field := reflect.Indirect(schema).FieldByName("MapOfInterface")
	s.True(field.IsValid())
}
