package tests

import kallax "github.com/src-d/go-kallax"

func (s *CommonSuite) TestQueryFindById() {
	store := NewResultSetFixtureStore(s.db)

	doc := NewResultSetFixture("bar")
	s.Nil(store.Insert(doc))

	query := NewResultSetFixtureQuery()
	query.Where(kallax.Eq(Schema.ResultSetFixture.ID, doc.ID))

	s.NotPanics(func() {
		s.Equal("bar", store.MustFindOne(query).Foo)
	})
}
