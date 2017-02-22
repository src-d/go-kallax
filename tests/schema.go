package tests

import "gopkg.in/src-d/go-kallax.v1"

type SchemaFixture struct {
	kallax.Model `table:"schema"`
	ID           kallax.ULID `pk:""`
	ShouldIgnore string      `kallax:"-"`
	String       string
	Int          int
	Nested       *SchemaFixture
	Inline       struct {
		Inline string
	} `kallax:",inline"`
	MapOfString    map[string]string
	MapOfInterface map[string]interface{}
	MapOfSomeType  map[string]struct {
		Foo string
	}
}

func newSchemaFixture() *SchemaFixture {
	return &SchemaFixture{ID: kallax.NewULID()}
}
