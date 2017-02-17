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
		`CREATE TABLE IF NOT EXISTS query (
			id uuid primary key,
			foo varchar(10)
		)`,
	}
	suite.Run(t, &QuerySuite{NewBaseSuite(schema, "query")})
}

func (s *QuerySuite) TestQuery() {
	store := NewQueryFixtureStore(s.db)
	doc := NewQueryFixture("bar")
	s.Nil(store.Insert(doc))

	query := NewQueryFixtureQuery()
	query.Where(kallax.Eq(Schema.QueryFixture.ID, doc.ID))

	s.NotPanics(func() {
		s.Equal("bar", store.MustFindOne(query).Foo)
	})

	notID := kallax.NewULID()
	queryErr := NewQueryFixtureQuery()
	queryErr.Where(kallax.Eq(Schema.QueryFixture.ID, notID))
	s.Panics(func() {
		s.Equal("bar", store.MustFindOne(queryErr).Foo)
	})
}
