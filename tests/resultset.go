package tests

import "gopkg.in/src-d/storable.v1"

type ResultSetFixture struct {
	storable.Document `bson:",inline" collection:"resultset"`
	Foo               string
}

func newResultSetFixture(f string) *ResultSetFixture {
	return &ResultSetFixture{Foo: f}
}

type ResultSetInitFixture struct {
	storable.Document `bson:",inline" collection:"resultset"`
	Foo               string
}

func (r *ResultSetInitFixture) Init(doc storable.DocumentBase) error {
	r.Foo = "foo"
	return nil
}
