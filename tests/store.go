package tests

import (
	"time"

	"github.com/src-d/go-kallax"
	"gopkg.in/mgo.v2/bson"
)

type StoreFixture struct {
	kallax.Document `bson:",inline" collection:"store"`
	Foo             string
}

type StoreWithConstructFixture struct {
	kallax.Document `bson:",inline" collection:"store_construct"`
	Foo             string
}

func newStoreWithConstructFixture(f string) *StoreWithConstructFixture {
	if f == "" {
		return nil
	}
	return &StoreWithConstructFixture{Foo: f}
}

type StoreWithNewFixture struct {
	kallax.Document `bson:",inline" collection:"store_new"`
	Foo             string
	Bar             string
}

func (s *StoreWithNewFixtureStore) New(f, b string) *StoreWithNewFixture {
	doc := &StoreWithNewFixture{Foo: f, Bar: b}

	doc.SetIsNew(true)
	doc.SetId(bson.NewObjectId())

	return doc
}

type MultiKeySortFixture struct {
	kallax.Document `bson:",inline" collection:"query"`
	Name            string
	Start           time.Time
	End             time.Time
}
