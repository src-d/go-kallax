package tests

import "github.com/src-d/go-kallax"

type QueryFixture struct {
	kallax.Model `table:"query"`
	ID           kallax.ULID `pk:""`
	Foo          string
}

func newQueryFixture(f string) *QueryFixture {
	return &QueryFixture{ID: kallax.NewULID(), Foo: f}
}
