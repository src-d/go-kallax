package tests

import "github.com/src-d/go-kallax"

type ResultSetFixture struct {
	kallax.Document `bson:",inline" collection:"resultset"`
	Foo             string
}

func newResultSetFixture(f string) *ResultSetFixture {
	return &ResultSetFixture{Foo: f}
}

type ResultSetInitFixture struct {
	kallax.Document `bson:",inline" collection:"resultset"`
	Foo             string
}

func (r *ResultSetInitFixture) Init(doc kallax.DocumentBase) error {
	r.Foo = "foo"
	return nil
}
