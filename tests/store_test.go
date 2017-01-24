package tests

import (
	"time"

	kallax "github.com/src-d/go-kallax"
)

func (s *CommonSuite) TestStoreNew() {
	doc := NewStoreFixture()
	s.False(doc.IsPersisted())
	s.Len(doc.ID.String(), 24)
}

func (s *CommonSuite) TestStoreQuery() {
	q := NewStoreFixtureQuery()
	s.NotNil(q)
}

func (s *CommonSuite) TestStoreFindAndCount() {
	store := NewStoreFixtureStore(s.db)
	s.Nil(store.Insert(NewStoreFixture()))
	s.Nil(store.Insert(NewStoreFixture()))

	query := NewStoreFixtureQuery()
	rs, err := store.Find(query)
	s.NotNil(rs)
	s.Nil(err)

	count, err := store.Count(query)
	s.Nil(err)
	s.Equal(2, count)
}

func (s *CommonSuite) TestStoreMustFind() {
	store := NewStoreFixtureStore(s.db)
	s.Nil(store.Insert(NewStoreFixture()))
	s.Nil(store.Insert(NewStoreFixture()))

	query := NewStoreFixtureQuery()
	s.NotPanics(func() {
		rs := store.MustFind(query)
		s.NotNil(rs)
	})

}

func (s *CommonSuite) TestStoreFailingOnNew() {
	doc := NewStoreWithConstructFixture("")
	s.Nil(doc)
}

func (s *CommonSuite) TestStoreFindOne() {
	store := NewStoreWithConstructFixtureStore(s.db)
	s.Nil(store.Insert(NewStoreWithConstructFixture("bar")))

	doc, err := store.FindOne(NewStoreWithConstructFixtureQuery())
	s.Nil(err)
	s.NotNil(doc)
	if err != nil {
		s.Nil(err, "This testcase was aborted")
		return
	}

	s.Equal("bar", doc.Foo)
}

func (s *CommonSuite) TestStoreMustFindOne() {
	store := NewStoreWithConstructFixtureStore(s.db)
	s.Nil(store.Insert(NewStoreWithConstructFixture("foo")))
	s.NotPanics(func() {
		s.Equal("foo", store.MustFindOne(NewStoreWithConstructFixtureQuery()).Foo)
	})
}

func (s *CommonSuite) TestStoreInsertUpdate() {
	store := NewStoreWithConstructFixtureStore(s.db)

	doc := NewStoreWithConstructFixture("foo")
	err := store.Insert(doc)
	s.Nil(err)
	s.NotPanics(func() {
		s.Equal("foo", store.MustFindOne(NewStoreWithConstructFixtureQuery()).Foo)
	})

	doc.Foo = "bar"
	updatedRows, err := store.Update(doc)
	s.Nil(err)
	s.True(updatedRows > 0)
	s.NotPanics(func() {
		s.Equal("bar", store.MustFindOne(NewStoreWithConstructFixtureQuery()).Foo)
	})
}

func (s *CommonSuite) TestStoreSave() {
	store := NewStoreWithConstructFixtureStore(s.db)

	doc := NewStoreWithConstructFixture("foo")
	updated, err := store.Save(doc)
	s.Nil(err)
	s.False(updated)
	s.True(doc.IsPersisted())
	s.NotPanics(func() {
		s.Equal("foo", store.MustFindOne(NewStoreWithConstructFixtureQuery()).Foo)
	})

	doc.Foo = "bar"
	updated, err = store.Save(doc)
	s.Nil(err)
	s.True(updated)
	s.NotPanics(func() {
		s.Equal("bar", store.MustFindOne(NewStoreWithConstructFixtureQuery()).Foo)
	})
}

func (s *CommonSuite) TestStoreCustomNew() {
	store := NewStoreWithNewFixtureStore(s.db)

	doc := store.New("foo", "bar")
	updated, err := store.Save(doc)
	s.Nil(err)
	s.False(updated)
	s.False(doc.IsPersisted())
	s.NotPanics(func() {
		s.Equal("foo", store.MustFindOne(NewStoreWithNewFixtureQuery()).Foo)
	})
	s.NotPanics(func() {
		s.Equal("bar", store.MustFindOne(NewStoreWithNewFixtureQuery()).Bar)
	})
}

func (s *CommonSuite) TestMultiKeySort() {
	store := NewMultiKeySortFixtureStore(s.db)

	var (
		doc *MultiKeySortFixture
		err error
	)

	doc = NewMultiKeySortFixture()
	doc.Name = "2015-2013"
	doc.Start = time.Date(2005, 1, 2, 0, 0, 0, 0, time.UTC)
	doc.End = time.Date(2013, 1, 2, 0, 0, 0, 0, time.UTC)
	err = store.Insert(doc)
	s.Nil(err)

	doc = NewMultiKeySortFixture()
	doc.Name = "2015-2012"
	doc.Start = time.Date(2005, 1, 2, 0, 0, 0, 0, time.UTC)
	doc.End = time.Date(2012, 4, 5, 0, 0, 0, 0, time.UTC)
	err = store.Insert(doc)
	s.Nil(err)

	doc = NewMultiKeySortFixture()
	doc.Name = "2002-2012"
	doc.Start = time.Date(2002, 1, 2, 0, 0, 0, 0, time.UTC)
	doc.End = time.Date(2012, 1, 2, 0, 0, 0, 0, time.UTC)
	err = store.Insert(doc)
	s.Nil(err)

	doc = NewMultiKeySortFixture()
	doc.Name = "2001-2012"
	doc.Start = time.Date(2001, 1, 2, 0, 0, 0, 0, time.UTC)
	doc.End = time.Date(2012, 1, 2, 0, 0, 0, 0, time.UTC)
	err = store.Insert(doc)
	s.Nil(err)

	q := NewMultiKeySortFixtureQuery()
	q.Order(kallax.Desc(Schema.MultiKeySortFixture.End), kallax.Desc(Schema.MultiKeySortFixture.Start))

	set, err := store.Find(q)
	s.Nil(err)

	if err != nil {
		s.Nil(err, "This testcase was aborted")
		return
	}

	documents, err := set.All()
	s.Nil(err)

	s.Len(documents, 4)
	s.Equal("2015-2013", documents[0].Name)
	s.Equal("2015-2012", documents[1].Name)
	s.Equal("2002-2012", documents[2].Name)
	s.Equal("2001-2012", documents[3].Name)
}
