package tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/T-M-A/go-kallax"
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
		rs := store.MustFind(NewResultSetFixtureQuery())
		docs, err := rs.All()
		s.Nil(err)
		s.Len(docs, 2)
	})
}

func (s *ResulsetSuite) TestResultSetOne() {
	store := NewResultSetFixtureStore(s.db)
	s.Nil(store.Insert(NewResultSetFixture("bar")))

	s.NotPanics(func() {
		rs := store.MustFind(NewResultSetFixtureQuery())
		doc, err := rs.One()
		s.Nil(err)
		s.Equal("bar", doc.Foo)
	})
}

func (s *ResulsetSuite) TestResultSetNextEmpty() {
	store := NewResultSetFixtureStore(s.db)

	s.NotPanics(func() {
		rs := store.MustFind(NewResultSetFixtureQuery())
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
		rs := store.MustFind(NewResultSetFixtureQuery())
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

func (s *ResulsetSuite) TestResultSetForEach() {
	store := NewResultSetFixtureStore(s.db)
	s.Nil(store.Insert(NewResultSetFixture("bar")))
	s.Nil(store.Insert(NewResultSetFixture("foo")))

	s.NotPanics(func() {
		count := 0
		rs := store.MustFind(NewResultSetFixtureQuery())
		err := rs.ForEach(func(*ResultSetFixture) error {
			count++
			return nil
		})

		s.Nil(err)
		s.Equal(2, count)
	})
}

func (s *ResulsetSuite) TestResultSetForEachStop() {
	store := NewResultSetFixtureStore(s.db)
	s.Nil(store.Insert(NewResultSetFixture("bar")))
	s.Nil(store.Insert(NewResultSetFixture("foo")))

	s.NotPanics(func() {
		count := 0
		rs := store.MustFind(NewResultSetFixtureQuery())
		err := rs.ForEach(func(*ResultSetFixture) error {
			count++
			return kallax.ErrStop
		})

		s.Nil(err)
		s.Equal(1, count)
	})
}

func (s *ResulsetSuite) TestResultSetForEachError() {
	store := NewResultSetFixtureStore(s.db)
	s.Nil(store.Insert(NewResultSetFixture("bar")))
	s.Nil(store.Insert(NewResultSetFixture("foo")))

	fail := errors.New("kallax: foo")

	s.NotPanics(func() {
		rs := store.MustFind(NewResultSetFixtureQuery())
		defer rs.Close()
		err := rs.ForEach(func(*ResultSetFixture) error {
			return fail
		})

		s.Equal(fail, err)
	})
}

func (s *ResulsetSuite) TestForEachAndCount() {
	store := NewResultSetFixtureStore(s.db)

	docInserted1 := NewResultSetFixture("bar")
	s.Nil(store.Insert(docInserted1))
	docInserted2 := NewResultSetFixture("baz")
	s.Nil(store.Insert(docInserted2))

	query := NewResultSetFixtureQuery()
	rs, err := store.Find(query)
	s.Nil(err)
	manualCount := 0
	rs.ForEach(func(doc *ResultSetFixture) error {
		manualCount++
		s.NotNil(doc)
		return nil
	})
	s.Equal(2, manualCount)

	queriedCount, err := store.Count(query)
	s.NoError(err)
	s.Equal(int64(2), queriedCount)
}
