package tests

import "gopkg.in/src-d/storable.v1"

type QueryFixture struct {
	storable.Document `bson:",inline" collection:"query"`
	Foo               string
}

func newQueryFixture(f string) *QueryFixture {
	return &QueryFixture{Foo: f}
}
