package tests

import "github.com/src-d/go-kallax"

type SchemaFixture struct {
	kallax.Model `table:"schema"`

	String string
	Int    int
	Nested *SchemaFixture
	Inline struct {
		Inline string
	} `kallax:",inline"`
	MapOfString    map[string]string
	MapOfInterface map[string]interface{}
	MapOfSomeType  map[string]struct {
		Foo string
	}
}
