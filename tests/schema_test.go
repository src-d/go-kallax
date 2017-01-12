package tests

import . "gopkg.in/check.v1"

func (s *MongoSuite) TestSchemaBasicField(c *C) {
	c.Assert(Schema.SchemaFixture.String.String(), Equals, "string")
}

func (s *MongoSuite) TestSchemaRanamedField(c *C) {
	c.Assert(Schema.SchemaFixture.Int.String(), Equals, "foo")
}

func (s *MongoSuite) TestSchemaInlineField(c *C) {
	c.Assert(Schema.SchemaFixture.Inline.Inline.String(), Equals, "inline")
}

func (s *MongoSuite) TestSchemaNestedField(c *C) {
	c.Assert(Schema.SchemaFixture.Nested.Int.String(), Equals, "nested.foo")
}

func (s *MongoSuite) TestSchemaMapsOfString(c *C) {
	key := Schema.SchemaFixture.MapOfString.Key("foo").String()
	c.Assert(key, Equals, "mapofstring.foo")
}

func (s *MongoSuite) TestSchemaMapOfSomeType(c *C) {
	key := Schema.SchemaFixture.MapOfSomeType.Foo.Key("qux").String()
	c.Assert(key, Equals, "mapofsometype.qux.foo")
}

func (s *MongoSuite) TestSchemaMapOfInterface(c *C) {
	key := Schema.SchemaFixture.MapOfInterface.Key("foo").String()
	c.Assert(key, Equals, "mapofinterface.foo")
}
