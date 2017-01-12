package tests

import (
	"errors"

	. "gopkg.in/check.v1"
)

func (s *MongoSuite) TestEventsInsert(c *C) {
	store := NewEventsFixtureStore(s.db)

	doc := store.New()
	err := store.Insert(doc)
	c.Assert(err, IsNil)
	c.Assert(doc.Checks, DeepEquals, map[string]bool{
		"BeforeInsert": true,
		"AfterInsert":  true,
	})
}

func (s *MongoSuite) TestEventsUpdate(c *C) {
	store := NewEventsFixtureStore(s.db)

	doc := store.New()
	err := store.Insert(doc)
	c.Assert(err, IsNil)

	doc.Checks = make(map[string]bool, 0)
	err = store.Update(doc)
	c.Assert(err, IsNil)
	c.Assert(doc.Checks, DeepEquals, map[string]bool{
		"BeforeUpdate": true,
		"AfterUpdate":  true,
	})
}

func (s *MongoSuite) TestEventsUpdateError(c *C) {
	store := NewEventsFixtureStore(s.db)

	doc := store.New()
	err := store.Insert(doc)
	doc.Checks = make(map[string]bool, 0)

	doc.MustFailAfter = errors.New("after")
	err = store.Update(doc)
	c.Assert(err, Equals, doc.MustFailAfter)

	doc.MustFailBefore = errors.New("before")
	err = store.Update(doc)
	c.Assert(err, Equals, doc.MustFailBefore)
}

func (s *MongoSuite) TestEventsSaveOnInsert(c *C) {
	store := NewEventsFixtureStore(s.db)

	doc := store.New()
	updated, err := store.Save(doc)
	c.Assert(err, IsNil)
	c.Assert(updated, Equals, false)
	c.Assert(doc.Checks, DeepEquals, map[string]bool{
		"BeforeInsert": true,
		"AfterInsert":  true,
	})
}

func (s *MongoSuite) TestEventsSaveOnUpdate(c *C) {
	store := NewEventsFixtureStore(s.db)

	doc := store.New()
	err := store.Insert(doc)
	doc.Checks = make(map[string]bool, 0)

	updated, err := store.Save(doc)
	c.Assert(err, IsNil)
	c.Assert(updated, Equals, true)
	c.Assert(doc.Checks, DeepEquals, map[string]bool{
		"BeforeUpdate": true,
		"AfterUpdate":  true,
	})
}

func (s *MongoSuite) TestEventsSaveInsert(c *C) {
	store := NewEventsSaveFixtureStore(s.db)

	doc := store.New()
	err := store.Insert(doc)
	c.Assert(err, IsNil)
	c.Assert(doc.Checks, DeepEquals, map[string]bool{
		"BeforeSave": true,
		"AfterSave":  true,
	})
}

func (s *MongoSuite) TestEventsSaveUpdate(c *C) {
	store := NewEventsSaveFixtureStore(s.db)

	doc := store.New()
	err := store.Insert(doc)
	c.Assert(err, IsNil)

	doc.Checks = make(map[string]bool, 0)
	err = store.Update(doc)
	c.Assert(err, IsNil)
	c.Assert(doc.Checks, DeepEquals, map[string]bool{
		"BeforeSave": true,
		"AfterSave":  true,
	})
}

func (s *MongoSuite) TestEventsSaveSave(c *C) {
	store := NewEventsSaveFixtureStore(s.db)

	doc := store.New()
	err := store.Insert(doc)
	doc.Checks = map[string]bool{"AfterInsert": true}

	updated, err := store.Save(doc)
	c.Assert(err, IsNil)
	c.Assert(updated, Equals, true)
	c.Assert(doc.Checks, DeepEquals, map[string]bool{
		"AfterInsert": true,
		"BeforeSave":  true,
		"AfterSave":   true,
	})
}
