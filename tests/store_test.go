package tests

import (
	"testing"
	"time"

	kallax "github.com/src-d/go-kallax"
	"github.com/stretchr/testify/suite"
)

func TestStoreSuite(t *testing.T) {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS store_construct (
			id uuid primary key,
			foo varchar(10)
		)`,
		`CREATE TABLE IF NOT EXISTS store (
			id uuid primary key,
			foo varchar(10),
			slice_prop text[],
			alias_slice_prop text[]
		)`,
		`CREATE TABLE IF NOT EXISTS store_new (
			id uuid primary key,
			foo varchar(10),
			bar varchar(10)
		)`,
		`CREATE TABLE IF NOT EXISTS query (
			id uuid primary key,
			name varchar(10),
			start timestamp,
			_end timestamp
		)`,
	}
	suite.Run(t, &StoreSuite{NewBaseSuite(schema, "store_construct", "store", "store_new", "query")})
}

type StoreSuite struct {
	BaseTestSuite
}

func (s *StoreSuite) TestStoreNew() {
	doc := NewStoreWithConstructFixture("foo")
	s.False(doc.IsPersisted())
	s.False(doc.ID.IsEmpty())
	s.Equal("foo", doc.Foo)
}

func (s *StoreSuite) TestStoreQuery() {
	q := NewStoreFixtureQuery()
	s.NotNil(q)
}

func (s *StoreSuite) TestStoreMustFind() {
	store := NewStoreFixtureStore(s.db)
	s.Nil(store.Insert(NewStoreFixture()))
	s.Nil(store.Insert(NewStoreFixture()))

	query := NewStoreFixtureQuery()
	s.NotPanics(func() {
		rs := store.MustFind(query)
		s.NotNil(rs)
	})
}

func (s *StoreSuite) TestStoreFailingOnNew() {
	doc := NewStoreWithConstructFixture("")
	s.Nil(doc)
}

func (s *StoreSuite) TestStoreFindOneReturnValues() {
	store := NewStoreWithConstructFixtureStore(s.db)
	s.Nil(store.Insert(NewStoreWithConstructFixture("bar")))

	notFoundQuery := NewStoreWithConstructFixtureQuery()
	notFoundQuery.Where(kallax.Eq(Schema.ResultSetFixture.ID, kallax.NewULID()))
	doc, err := store.FindOne(notFoundQuery)
	s.resultOrError(doc, err)

	doc, err = store.FindOne(NewStoreWithConstructFixtureQuery())
	s.resultOrError(doc, err)
}

func (s *StoreSuite) TestStoreInsertUpdateMustFind() {
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

func (s *StoreSuite) TestStoreSave() {
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

func (s *StoreSuite) TestMultiKeySort() {
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
	if set == nil {
		s.Nil(err, "This testcase was aborted. ResultSet should not be nil")
		return
	}

	documents, err := set.All()
	s.Nil(err)
	s.Len(documents, 4)
	success := true
	for _, doc := range documents {
		if !s.NotNil(doc) {
			success = false
		}
	}

	if !success {
		s.Fail("Testcase aborted. All retrieved Documents should be not-nil")
		return
	}

	s.Equal("2015-2013", documents[0].Name)
	s.Equal("2015-2012", documents[1].Name)
	s.Equal("2002-2012", documents[2].Name)
	s.Equal("2001-2012", documents[3].Name)
}

func (s *StoreSuite) TestFindOne() {
	store := NewStoreWithConstructFixtureStore(s.db)

	docInserted := NewStoreWithConstructFixture("bar")
	s.Nil(store.Insert(docInserted))

	query := NewStoreWithConstructFixtureQuery()
	docFound, err := store.FindOne(query)
	s.resultOrError(docFound, err)
	if s.NotNil(docFound) {
		s.Equal(docInserted.Foo, docFound.Foo)
	}
}

func (s *StoreSuite) TestFindAliasSlice() {
	store := NewStoreFixtureStore(s.db)

	fixture1 := NewStoreFixture()
	fixture1.Foo = "ONE"
	s.Nil(store.Insert(fixture1))
	s.assertMustFindByFoo(store, "ONE")

	fixture2 := NewStoreFixture()
	fixture2.Foo = "TWO"
	fixture2.SliceProp = []string{"1", "2"}
	s.Nil(store.Insert(fixture2))
	s.assertMustFindByFoo(store, "TWO")

	fixture3 := NewStoreFixture()
	fixture3.Foo = "THREE"
	fixture3.AliasSliceProp = AliasSliceString{"1", "2"}
	s.Nil(store.Insert(fixture3))
	s.assertMustFindByFoo(store, "THREE")
}

func (s *StoreSuite) assertMustFindByFoo(st *StoreFixtureStore, foo string) {
	s.NotPanics(func() {
		q := NewStoreFixtureQuery().Where(kallax.Eq(Schema.StoreFixture.Foo, foo))
		r := st.MustFindOne(q)
		s.Equal(foo, r.Foo)
	})
}
