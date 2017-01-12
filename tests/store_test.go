package tests

import (
	"time"

	"github.com/src-d/go-kallax"
	. "gopkg.in/check.v1"
)

func (s *MongoSuite) TestStoreNew(c *C) {
	store := NewStoreFixtureStore(s.db)
	doc := store.New()

	c.Assert(doc.IsNew(), Equals, true)
	c.Assert(doc.GetId().Hex(), HasLen, 24)
}

func (s *MongoSuite) TestStoreQuery(c *C) {
	store := NewStoreFixtureStore(s.db)
	q := store.Query()
	c.Assert(q, Not(IsNil))
}

func (s *MongoSuite) TestStoreFind(c *C) {
	store := NewStoreFixtureStore(s.db)
	c.Assert(store.Insert(store.New()), IsNil)
	c.Assert(store.Insert(store.New()), IsNil)

	rs, err := store.Find(store.Query())
	c.Assert(err, IsNil)

	count, err := rs.Count()
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 2)
}

func (s *MongoSuite) TestStoreCount(c *C) {
	store := NewStoreFixtureStore(s.db)
	c.Assert(store.Insert(store.New()), IsNil)
	c.Assert(store.Insert(store.New()), IsNil)

	count, err := store.Count(store.Query())
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 2)
}

func (s *MongoSuite) TestStoreMustFind(c *C) {
	store := NewStoreFixtureStore(s.db)
	c.Assert(store.Insert(store.New()), IsNil)
	c.Assert(store.Insert(store.New()), IsNil)

	count, err := store.MustFind(store.Query()).Count()
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 2)
}

func (s *MongoSuite) TestStoreFailingOnNew(c *C) {
	store := NewStoreWithConstructFixtureStore(s.db)

	doc := store.New("")
	c.Assert(doc, IsNil)
}

func (s *MongoSuite) TestStoreFindOne(c *C) {
	store := NewStoreWithConstructFixtureStore(s.db)
	c.Assert(store.Insert(store.New("bar")), IsNil)

	doc, err := store.FindOne(store.Query())
	c.Assert(err, IsNil)
	c.Assert(doc.Foo, Equals, "bar")
}

func (s *MongoSuite) TestStoreMustFindOne(c *C) {
	store := NewStoreWithConstructFixtureStore(s.db)
	c.Assert(store.Insert(store.New("foo")), IsNil)
	c.Assert(store.MustFindOne(store.Query()).Foo, Equals, "foo")
}

func (s *MongoSuite) TestStoreInsertUpdate(c *C) {
	store := NewStoreWithConstructFixtureStore(s.db)

	doc := store.New("foo")
	err := store.Insert(doc)
	c.Assert(err, IsNil)
	c.Assert(store.MustFindOne(store.Query()).Foo, Equals, "foo")

	doc.Foo = "bar"
	err = store.Update(doc)
	c.Assert(err, IsNil)
	c.Assert(store.MustFindOne(store.Query()).Foo, Equals, "bar")
}

func (s *MongoSuite) TestStoreSave(c *C) {
	store := NewStoreWithConstructFixtureStore(s.db)

	doc := store.New("foo")
	updated, err := store.Save(doc)
	c.Assert(err, IsNil)
	c.Assert(updated, Equals, false)
	c.Assert(doc.IsNew(), Equals, false)
	c.Assert(store.MustFindOne(store.Query()).Foo, Equals, "foo")

	doc.Foo = "bar"
	updated, err = store.Save(doc)
	c.Assert(err, IsNil)
	c.Assert(updated, Equals, true)
	c.Assert(store.MustFindOne(store.Query()).Foo, Equals, "bar")
}

func (s *MongoSuite) TestStoreCustomNew(c *C) {
	store := NewStoreWithNewFixtureStore(s.db)

	doc := store.New("foo", "bar")
	updated, err := store.Save(doc)
	c.Assert(err, IsNil)
	c.Assert(updated, Equals, false)
	c.Assert(doc.IsNew(), Equals, false)
	c.Assert(store.MustFindOne(store.Query()).Foo, Equals, "foo")
	c.Assert(store.MustFindOne(store.Query()).Bar, Equals, "bar")
}

func (s *MongoSuite) TestMultiKeySort(c *C) {
	store := NewMultiKeySortFixtureStore(s.db)

	var (
		doc *MultiKeySortFixture
		err error
	)

	doc = store.New()
	doc.Name = "2015-2013"
	doc.Start = time.Date(2005, 1, 2, 0, 0, 0, 0, time.UTC)
	doc.End = time.Date(2013, 1, 2, 0, 0, 0, 0, time.UTC)
	err = store.Insert(doc)
	c.Assert(err, IsNil)

	doc = store.New()
	doc.Name = "2015-2012"
	doc.Start = time.Date(2005, 1, 2, 0, 0, 0, 0, time.UTC)
	doc.End = time.Date(2012, 4, 5, 0, 0, 0, 0, time.UTC)
	err = store.Insert(doc)
	c.Assert(err, IsNil)

	doc = store.New()
	doc.Name = "2002-2012"
	doc.Start = time.Date(2002, 1, 2, 0, 0, 0, 0, time.UTC)
	doc.End = time.Date(2012, 1, 2, 0, 0, 0, 0, time.UTC)
	err = store.Insert(doc)
	c.Assert(err, IsNil)

	doc = store.New()
	doc.Name = "2001-2012"
	doc.Start = time.Date(2001, 1, 2, 0, 0, 0, 0, time.UTC)
	doc.End = time.Date(2012, 1, 2, 0, 0, 0, 0, time.UTC)
	err = store.Insert(doc)
	c.Assert(err, IsNil)

	q := store.Query()
	q.Sort(kallax.Sort{
		kallax.FieldSort{Schema.MultiKeySortFixture.End, kallax.Desc},
		kallax.FieldSort{Schema.MultiKeySortFixture.Start, kallax.Desc},
	})

	set, err := store.Find(q)
	c.Assert(err, IsNil)

	documents, err := set.All()
	c.Assert(err, IsNil)

	c.Assert(documents, HasLen, 4)
	c.Assert(documents[0].Name, Equals, "2015-2013")
	c.Assert(documents[1].Name, Equals, "2015-2012")
	c.Assert(documents[2].Name, Equals, "2002-2012")
	c.Assert(documents[3].Name, Equals, "2001-2012")
}
