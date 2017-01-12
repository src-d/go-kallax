package tests

import "github.com/src-d/go-kallax"

type QueryFixture struct {
	kallax.Document `bson:",inline" collection:"query"`
	Foo             string
}

func newQueryFixture(f string) *QueryFixture {
	return &QueryFixture{Foo: f}
}
