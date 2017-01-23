package tests

import "github.com/src-d/go-kallax"

type QueryFixture struct {
	kallax.Model `table:"query"`
	Foo          string
}

func newQueryFixture(f string) *QueryFixture {
	return &QueryFixture{Foo: f}
}
