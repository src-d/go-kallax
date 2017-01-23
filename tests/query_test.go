package tests

import (
	"database/sql"

	"github.com/src-d/go-kallax"
	"github.com/stretchr/testify/suite"
)

type QuerySuite struct {
	suite.Suite
	db *sql.DB
}

func (s *QuerySuite) TestQueryFindById() {
	store := NewResultSetFixtureStore(s.db)

	doc := NewResultSetFixture("bar")
	s.Nil(store.Insert(doc))

	query := NewResultSetFixtureQuery()
	query.Where(kallax.Eq(Schema.ResultSetFixture.ID, doc.ID))

	s.NotPanics(func() {
		s.Equal("bar", store.MustFindOne(query).Foo)
	})
}
