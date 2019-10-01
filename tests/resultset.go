package tests

import "github.com/zbyte/go-kallax"

type ResultSetFixture struct {
	kallax.Model `table:"resultset"`
	ID           kallax.ULID `pk:""`
	Foo          string
}

func newResultSetFixture(f string) *ResultSetFixture {
	return &ResultSetFixture{ID: kallax.NewULID(), Foo: f}
}
