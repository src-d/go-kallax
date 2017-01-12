package tests

import "gopkg.in/src-d/storable.v1"

type SchemaFixture struct {
	storable.Document `bson:",inline" collection:"schema"`

	String string
	Int    int `bson:"foo"`
	Nested *SchemaFixture
	Inline struct {
		Inline string
	} `bson:",inline"`
	MapOfString    map[string]string
	MapOfInterface map[string]interface{}
	MapOfSomeType  map[string]struct {
		Foo string
	}
}
