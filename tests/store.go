package tests

import (
	"time"

	"github.com/src-d/go-kallax"
)

type StoreFixture struct {
	kallax.Model `table:"store"`
	Foo          string
}

type StoreWithConstructFixture struct {
	kallax.Model `table:"store_construct"`
	Foo          string
}

func newStoreWithConstructFixture(f string) *StoreWithConstructFixture {
	if f == "" {
		return nil
	}
	return &StoreWithConstructFixture{Foo: f}
}

type StoreWithNewFixture struct {
	kallax.Model `table:"store_new"`
	Foo          string
	Bar          string
}

func (s *StoreWithNewFixtureStore) New(f, b string) *StoreWithNewFixture {
	doc := &StoreWithNewFixture{Foo: f, Bar: b}
	doc.SetID(kallax.NewID())
	return doc
}

type MultiKeySortFixture struct {
	kallax.Model `table:"query"`
	Name         string
	Start        time.Time
	End          time.Time
}
