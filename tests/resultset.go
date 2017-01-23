package tests

import "github.com/src-d/go-kallax"

type ResultSetFixture struct {
	kallax.Model `table:"resultset"`
	Foo          string
}

func newResultSetFixture(f string) *ResultSetFixture {
	return &ResultSetFixture{Foo: f}
}

type ResultSetInitFixture struct {
	kallax.Model `table:"resultset"`
	Foo          string
}

func (r *ResultSetInitFixture) Init(doc kallax.Record) error {
	r.Foo = "foo"
	return nil
}
