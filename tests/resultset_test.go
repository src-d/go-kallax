package tests

import (
	"errors"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/storable.v1"
)

func (s *MongoSuite) TestResultSetAll(c *C) {
	store := NewResultSetFixtureStore(s.db)
	c.Assert(store.Insert(store.New("bar")), IsNil)
	c.Assert(store.Insert(store.New("foo")), IsNil)

	docs, err := store.MustFind(store.Query()).All()
	c.Assert(err, IsNil)
	c.Assert(docs, HasLen, 2)
}

func (s *MongoSuite) TestResultSetAllInit(c *C) {
	store := NewResultSetInitFixtureStore(s.db)

	c.Assert(store.Insert(store.New()), IsNil)
	c.Assert(store.Insert(store.New()), IsNil)

	docs, err := store.MustFind(store.Query()).All()
	c.Assert(err, IsNil)
	c.Assert(docs, HasLen, 2)
	c.Assert(docs[0].Foo, Equals, "foo")
	c.Assert(docs[1].Foo, Equals, "foo")
}

func (s *MongoSuite) TestResultSetOne(c *C) {
	store := NewResultSetFixtureStore(s.db)
	c.Assert(store.Insert(store.New("bar")), IsNil)

	doc, err := store.MustFind(store.Query()).One()
	c.Assert(err, IsNil)
	c.Assert(doc.Foo, Equals, "bar")
}

func (s *MongoSuite) TestResultInitSetOne(c *C) {
	store := NewResultSetInitFixtureStore(s.db)

	a := store.New()
	a.Foo = "qux"

	c.Assert(store.Insert(a), IsNil)

	doc, err := store.MustFind(store.Query()).One()
	c.Assert(err, IsNil)
	c.Assert(doc.Foo, Equals, "foo")
}

func (s *MongoSuite) TestResultSetNextEmpty(c *C) {
	store := NewResultSetFixtureStore(s.db)
	rs := store.MustFind(store.Query())
	returned := rs.Next()
	c.Assert(returned, Equals, false)

	doc, err := rs.Get()
	c.Assert(err, IsNil)
	c.Assert(doc, IsNil)
}

func (s *MongoSuite) TestResultSetNext(c *C) {
	store := NewResultSetFixtureStore(s.db)
	c.Assert(store.Insert(store.New("bar")), IsNil)

	rs := store.MustFind(store.Query())
	returned := rs.Next()
	c.Assert(returned, Equals, true)

	doc, err := rs.Get()
	c.Assert(err, IsNil)
	c.Assert(doc.Foo, Equals, "bar")

	returned = rs.Next()
	c.Assert(returned, Equals, false)

	doc, err = rs.Get()
	c.Assert(err, IsNil)
	c.Assert(doc, IsNil)
}

func (s *MongoSuite) TestResultSetInitNext(c *C) {
	store := NewResultSetInitFixtureStore(s.db)
	c.Assert(store.Insert(store.New()), IsNil)

	rs := store.MustFind(store.Query())
	returned := rs.Next()
	c.Assert(returned, Equals, true)

	doc, err := rs.Get()
	c.Assert(err, IsNil)
	c.Assert(doc.Foo, Equals, "foo")

	returned = rs.Next()
	c.Assert(returned, Equals, false)
}

func (s *MongoSuite) TestResultSetForEach(c *C) {
	store := NewResultSetFixtureStore(s.db)
	c.Assert(store.Insert(store.New("bar")), IsNil)
	c.Assert(store.Insert(store.New("foo")), IsNil)

	count := 0
	err := store.MustFind(store.Query()).ForEach(func(*ResultSetFixture) error {
		count++
		return nil
	})

	c.Assert(err, IsNil)
	c.Assert(count, Equals, 2)
}

func (s *MongoSuite) TestResultSetInitForEach(c *C) {
	store := NewResultSetInitFixtureStore(s.db)
	c.Assert(store.Insert(store.New()), IsNil)
	c.Assert(store.Insert(store.New()), IsNil)

	count := 0
	err := store.MustFind(store.Query()).ForEach(func(r *ResultSetInitFixture) error {
		c.Assert(r, NotNil)
		c.Assert(r.Foo, Equals, "foo")
		count++
		return nil
	})

	c.Assert(err, IsNil)
	c.Assert(count, Equals, 2)
}

func (s *MongoSuite) TestResultSetForEachStop(c *C) {
	store := NewResultSetFixtureStore(s.db)
	c.Assert(store.Insert(store.New("bar")), IsNil)
	c.Assert(store.Insert(store.New("foo")), IsNil)

	count := 0
	err := store.MustFind(store.Query()).ForEach(func(*ResultSetFixture) error {
		count++
		return storable.ErrStop
	})

	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)
}

func (s *MongoSuite) TestResultSetForEachError(c *C) {
	store := NewResultSetFixtureStore(s.db)
	c.Assert(store.Insert(store.New("bar")), IsNil)
	c.Assert(store.Insert(store.New("foo")), IsNil)

	fail := errors.New("foo")
	err := store.MustFind(store.Query()).ForEach(func(*ResultSetFixture) error {
		return fail
	})

	c.Assert(err, Equals, fail)
}
