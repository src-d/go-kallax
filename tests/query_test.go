package tests

import (
	"testing"

	"github.com/src-d/go-kallax"
	"github.com/stretchr/testify/suite"
)

type QuerySuite struct {
	BaseTestSuite
}

func TestQuerySuite(t *testing.T) {
	schema := []string{
		`CREATE TABLE query (
			id uuid primary key,
			foo varchar(10)
		)`,
	}
	suite.Run(t, &QuerySuite{BaseTestSuite{initQueries: schema}})
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
