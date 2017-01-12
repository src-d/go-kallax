package tests

import . "gopkg.in/check.v1"

func (s *MongoSuite) TestQueryFindById(c *C) {
	store := NewResultSetFixtureStore(s.db)

	doc := store.New("bar")
	c.Assert(store.Insert(doc), IsNil)

	q := store.Query().FindById(doc.Id)
	c.Assert(store.MustFindOne(q).Foo, Equals, "bar")
}
