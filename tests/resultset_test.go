package tests

import (
	"errors"
	"testing"

	"github.com/src-d/go-kallax"
	"github.com/stretchr/testify/suite"
)

type ResulsetSuite struct {
	BaseTestSuite
}

func TestResulsetSuite(t *testing.T) {
	schema := []string{
		`CREATE TABLE resultset (
			id uuid primary key,
			foo varchar(10)
		)`,
	}
	suite.Run(t, &ResulsetSuite{BaseTestSuite{initQueries: schema}})
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

func (s *ResulsetSuite) TestResultSetAllInit() {
	store := NewResultSetInitFixtureStore(s.db)

	s.Nil(store.Insert(NewResultSetInitFixture()))
	s.Nil(store.Insert(NewResultSetInitFixture()))

	s.NotPanics(func() {
		rs := store.MustFind(NewResultSetInitFixtureQuery())
		docs, err := rs.All()
		s.Nil(err)
		s.Len(docs, 2)
		s.Equal("foo", docs[0].Foo)
		s.Equal("foo", docs[1].Foo)
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

func (s *ResulsetSuite) TestResultInitSetOne() {
	store := NewResultSetInitFixtureStore(s.db)

	a := NewResultSetInitFixture()
	a.Foo = "qux"

	s.Nil(store.Insert(a))

	s.NotPanics(func() {
		rs := store.MustFind(NewResultSetInitFixtureQuery())
		doc, err := rs.One()
		s.Nil(err)
		s.Equal("foo", doc.Foo)
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

func (s *ResulsetSuite) TestResultSetInitNext() {
	store := NewResultSetInitFixtureStore(s.db)
	s.Nil(store.Insert(NewResultSetInitFixture()))

	s.NotPanics(func() {
		rs := store.MustFind(NewResultSetInitFixtureQuery())
		returned := rs.Next()
		s.True(returned)

		doc, err := rs.Get()
		s.Nil(err)
		s.Equal("foo", doc.Foo)

		returned = rs.Next()
		s.False(returned)
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

func (s *ResulsetSuite) TestResultSetInitForEach() {
	store := NewResultSetInitFixtureStore(s.db)
	s.Nil(store.Insert(NewResultSetInitFixture()))
	s.Nil(store.Insert(NewResultSetInitFixture()))

	s.NotPanics(func() {
		count := 0
		rs := store.MustFind(NewResultSetInitFixtureQuery())
		err := rs.ForEach(func(r *ResultSetInitFixture) error {
			s.Nil(r)
			s.Equal("foo", r.Foo)
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

	fail := errors.New("foo")

	s.NotPanics(func() {
		rs := store.MustFind(NewResultSetFixtureQuery())
		err := rs.ForEach(func(*ResultSetFixture) error {
			return fail
		})

		s.Equal(fail, err)
	})
}
