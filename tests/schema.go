package tests

import "github.com/networkteam/go-kallax"

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
	Inverse *SchemaRelationshipFixture `fk:"rel_id,inverse"`
}

type SchemaRelationshipFixture struct {
	kallax.Model `table:"relationship"`
	ID           kallax.ULID `pk:""`
}

func newSchemaFixture() *SchemaFixture {
	return &SchemaFixture{ID: kallax.NewULID()}
}
