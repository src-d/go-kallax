package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ResulsetSuite struct {
	BaseTestSuite
}

func TestResulsetSuite(t *testing.T) {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS resultset (
			id uuid primary key,
			foo varchar(10)
		)`,
	}
	suite.Run(t, &ResulsetSuite{NewBaseSuite(schema, "resultset")})
}

func (s *ResulsetSuite) TestResultSetAll() {
	store := NewResultSetFixtureStore(s.db)
	s.Nil(store.Insert(NewResultSetFixture("bar")))
	s.Nil(store.Insert(NewResultSetFixture("foo")))

	s.NotPanics(func() {
		rs, err := store.Find(NewResultSetFixtureQuery())
		s.Nil(err)
		docs, err := rs.All()
		s.Nil(err)
		s.Len(docs, 2)
	})
}

func (s *ResulsetSuite) TestResultSetOne() {
	store := NewResultSetFixtureStore(s.db)
	s.Nil(store.Insert(NewResultSetFixture("bar")))

	s.NotPanics(func() {
		rs, err := store.Find(NewResultSetFixtureQuery())
		s.Nil(err)
		doc, err := rs.One()
		s.Nil(err)
		s.Equal("bar", doc.Foo)
	})
}

func (s *ResulsetSuite) TestResultSetNextEmpty() {
	store := NewResultSetFixtureStore(s.db)

	s.NotPanics(func() {
		rs, err := store.Find(NewResultSetFixtureQuery())
		s.Nil(err)
		returned := rs.Next()
		s.False(returned)

		doc, err := rs.Get()
		s.Nil(err)
		s.Nil(doc)
	})
}

func (s *ResulsetSuite) TestResultSetNext() {
	store := NewResultSetFixtureStore(s.db)
	s.Nil(store.Insert(NewResultSetFixture("bar")))

	s.NotPanics(func() {
		rs, err := store.Find(NewResultSetFixtureQuery())
		s.Nil(err)
		returned := rs.Next()
		s.True(returned)

		doc, err := rs.Get()
		s.Nil(err)
		s.Equal("bar", doc.Foo)

		returned = rs.Next()
		s.False(returned)

		doc, err = rs.Get()
		s.Nil(err)
		s.Nil(doc)
	})
}

func (s *ResulsetSuite) TestForEachAndCount() {
	store := NewResultSetFixtureStore(s.db)

	docInserted1 := NewResultSetFixture("bar")
	s.Nil(store.Insert(docInserted1))
	docInserted2 := NewResultSetFixture("baz")
	s.Nil(store.Insert(docInserted2))

	query := NewResultSetFixtureQuery()
	_, err := store.FindAll(query)
	s.Nil(err)

	queriedCount, err := store.Count(query)
	s.NoError(err)
	s.Equal(int64(2), queriedCount)
}
