package tests

import (
	"time"

	"gopkg.in/mgo.v2/bson"
	"gopkg.in/src-d/storable.v1"
)

type StoreFixture struct {
	storable.Document `bson:",inline" collection:"store"`
	Foo               string
}

type StoreWithConstructFixture struct {
	storable.Document `bson:",inline" collection:"store_construct"`
	Foo               string
}

func newStoreWithConstructFixture(f string) *StoreWithConstructFixture {
	if f == "" {
		return nil
	}
	return &StoreWithConstructFixture{Foo: f}
}

type StoreWithNewFixture struct {
	storable.Document `bson:",inline" collection:"store_new"`
	Foo               string
	Bar               string
}

func (s *StoreWithNewFixtureStore) New(f, b string) *StoreWithNewFixture {
	doc := &StoreWithNewFixture{Foo: f, Bar: b}

	doc.SetIsNew(true)
	doc.SetId(bson.NewObjectId())

	return doc
}

type MultiKeySortFixture struct {
	storable.Document `bson:",inline" collection:"query"`
	Name              string
	Start             time.Time
	End               time.Time
}
