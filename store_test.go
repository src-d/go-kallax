package kallax

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StoreSuite struct {
	suite.Suite
	db       *sql.DB
	store    *Store
	errDB    *sql.DB
	errStore *Store
}

func (s *StoreSuite) SetupTest() {
	var err error
	s.db, err = openTestDB()
	s.NoError(err)

	s.store = NewStore(s.db)
	setupTables(s.T(), s.db)

	s.errDB, err = sql.Open("postgres", "postgres://0.0.0.0:5432/notexists")
	s.NoError(err)
	s.errStore = NewStore(s.errDB)
}

func (s *StoreSuite) TearDownTest() {
	teardownTables(s.T(), s.db)
	s.NoError(s.db.Close())
	s.NoError(s.errDB.Close())
}

func (s *StoreSuite) TestInsert() {
	m := newModel("a", "a@a.a", 1)
	s.NoError(s.store.Insert(ModelSchema, m))
	s.True(m.IsPersisted(), "model should be persisted now")
	s.assertModel(m)
}

func (s *StoreSuite) TestInsert_Fail() {
	m := newModel("a", "a@a.a@", 1)
	s.Error(s.errStore.Insert(ModelSchema, m))
}

func (s *StoreSuite) TestInsert_NotNew() {
	var m model
	m.setPersisted()
	s.Equal(ErrNonNewDocument, s.store.Insert(ModelSchema, &m))
}

func (s *StoreSuite) TestInsert_IDEmpty() {
	var m = new(model)
	s.NoError(s.store.Insert(ModelSchema, m))
	s.False(m.GetID().IsEmpty())
}

func (s *StoreSuite) TestInsert_NoColumns() {
	var m = new(onlyPkModel)
	s.Equal(ErrNoColumns, s.store.Insert(onlyPkModelSchema, m))
}

func (s *StoreSuite) TestUpdate() {
	var m = newModel("a", "a@a.a", 1)
	s.NoError(s.store.Insert(ModelSchema, m))

	var newModel = newModel("a", "a@a.a", 1)
	newModel.ID = m.ID
	_, err := s.store.Update(ModelSchema, newModel)
	s.Equal(ErrNewDocument, err)

	newModel.setPersisted()
	newModel.ID = 0
	_, err = s.store.Update(ModelSchema, newModel)
	s.Equal(ErrEmptyID, err)

	m.Age = 2
	m.Email = "b@b.b"
	m.Name = "b"
	rows, err := s.store.Update(ModelSchema, m)
	s.NoError(err)
	s.Equal(int64(1), rows, "rows affected")
	s.assertModel(m)

	m.setWritable(false)
	_, err = s.store.Update(ModelSchema, m)
	s.Equal(ErrNotWritable, err)
}

func (s *StoreSuite) TestUpdate_ColumnNotFound() {
	var m = newModel("a", "a@a.a", 1)
	s.NoError(s.store.Insert(ModelSchema, m))

	_, err := s.store.Update(ModelSchema, m, f("not_exists"))
	s.Error(err)
}

func (s *StoreSuite) TestUpdate_NotUpdated() {
	var m = newModel("a", "a@a.a", 1)
	s.NoError(s.store.Insert(ModelSchema, m))

	m.ID = 567
	_, err := s.store.Update(ModelSchema, m)
	s.Equal(ErrNoRowUpdate, err)
}

func (s *StoreSuite) TestUpdate_Fail() {
	var m = newModel("a", "a@a.a", 1)
	m.setPersisted()

	_, err := s.errStore.Update(ModelSchema, m)
	s.Error(err)
}

func (s *StoreSuite) TestSave() {
	m := newModel("a", "a@a.a", 1)
	updated, err := s.store.Save(ModelSchema, m)
	s.NoError(err)
	s.False(updated)
	s.assertModel(m)

	m.Age = 5
	updated, err = s.store.Save(ModelSchema, m)
	s.NoError(err)
	s.True(updated)

	m.setWritable(false)
	_, err = s.store.Save(ModelSchema, m)
	s.Equal(ErrNotWritable, err)
}

func (s *StoreSuite) TestDelete() {
	m := newModel("a", "a@a.a", 1)
	s.NoError(s.store.Insert(ModelSchema, m))
	s.assertModel(m)

	s.NoError(s.store.Delete(ModelSchema, m))
	s.assertNotExists(m)

	var mod model
	s.Equal(ErrEmptyID, s.store.Delete(nil, &mod))
}

func (s *StoreSuite) TestRawQuery() {
	s.NoError(s.store.Insert(ModelSchema, newModel("Joe", "", 1)))
	s.NoError(s.store.Insert(ModelSchema, newModel("Jane", "", 2)))
	s.NoError(s.store.Insert(ModelSchema, newModel("Anna", "", 2)))

	rs, err := s.store.RawQuery("SELECT name FROM model WHERE age > $1", 1)
	s.NoError(err)

	var names []string
	for rs.Next() {
		_, err := rs.Get(ModelSchema)
		s.Equal(ErrRawScan, err)
		var name string
		s.NoError(rs.RawScan(&name))
		names = append(names, name)
	}
	s.Equal([]string{"Jane", "Anna"}, names)
}

func (s *StoreSuite) TestRawQuery_Transaction() {
	s.NoError(s.store.Transaction(func(store *Store) error {
		rs, err := store.RawQuery("SELECT 1 + 1")
		s.NoError(err)

		var n int
		rs.Next()
		s.NoError(rs.RawScan(&n))
		s.Equal(2, n)
		s.NoError(rs.Close())
		return nil
	}))
}

func (s *StoreSuite) TestRawQuery_Fail() {
	rs, err := s.errStore.RawQuery("SELECT name FROM model WHERE age > $1", 1)
	s.Nil(rs)
	s.Error(err)
}

func (s *StoreSuite) TestRawExec() {
	s.NoError(s.store.Insert(ModelSchema, newModel("Joe", "", 1)))
	s.NoError(s.store.Insert(ModelSchema, newModel("Jane", "", 2)))
	s.NoError(s.store.Insert(ModelSchema, newModel("Anna", "", 2)))

	rows, err := s.store.RawExec("DELETE FROM model WHERE age > $1", 1)
	s.NoError(err)
	s.Equal(int64(2), rows)
}

func (s *StoreSuite) TestRawExec_Fail() {
	rows, err := s.errStore.RawExec("DELETE FROM model WHERE age > $1", 1)
	s.Equal(int64(0), rows)
	s.Error(err)
}

func (s *StoreSuite) TestFind() {
	s.NoError(s.store.Insert(ModelSchema, newModel("Joe", "", 1)))
	s.NoError(s.store.Insert(ModelSchema, newModel("Jane", "", 2)))
	s.NoError(s.store.Insert(ModelSchema, newModel("Anna", "", 2)))

	q := NewBaseQuery(ModelSchema)
	q.Select(f("name"))
	q.Where(Gt(f("age"), 1))

	rs, err := s.store.Find(q)
	s.NoError(err)
	s.assertFound(rs, "Jane", "Anna")

	q = NewBaseQuery(ModelSchema)
	q.Select(f("name"))
	q.Limit(1)
	q.Offset(1)

	rs, err = s.store.Find(q)
	s.NoError(err)
	s.assertFound(rs, "Jane")
}

func (s *StoreSuite) TestFind_Fail() {
	q := NewBaseQuery(ModelSchema)
	_, err := s.errStore.Find(q)
	s.Error(err)
}

func (s *StoreSuite) TestMustFind() {
	s.NoError(s.store.Insert(ModelSchema, newModel("Joe", "", 1)))

	q := NewBaseQuery(ModelSchema)
	s.NotPanics(func() {
		rs := s.store.MustFind(q)
		s.assertFound(rs, "Joe")
	})

	s.Panics(func() {
		s.errStore.Debug().MustFind(q)
	})
}

func (s *StoreSuite) TestDebugWith() {
	var queries []string
	var logger = func(q string, args ...interface{}) {
		queries = append(queries, q)
	}
	s.store.DebugWith(logger).RawQuery("SELECT 1 + 1")
	s.store.DebugWith(logger).RawExec("UPDATE foo SET bar = 1")

	s.Equal(
		queries,
		[]string{
			"kallax: Query: SELECT 1 + 1",
			"kallax: Exec: UPDATE foo SET bar = 1",
		},
	)
}

func (s *StoreSuite) assertFound(rs ResultSet, expected ...string) {
	var names []string
	for rs.Next() {
		record, err := rs.Get(ModelSchema)
		s.NoError(err)
		m, ok := record.(*model)
		s.True(ok)
		s.True(m.IsPersisted())
		names = append(names, m.Name)
	}
	s.Equal(expected, names)
}

func (s *StoreSuite) TestCount() {
	s.NoError(s.store.Insert(ModelSchema, newModel("Joe", "", 1)))
	s.NoError(s.store.Insert(ModelSchema, newModel("Jane", "", 2)))
	s.NoError(s.store.Insert(ModelSchema, newModel("Anna", "", 2)))

	q := NewBaseQuery(ModelSchema)
	q.Select(f("name"))
	q.Where(Gt(f("age"), 1))

	cnt, err := s.store.Count(q)
	s.NoError(err)
	s.Equal(int64(2), cnt)
}

func (s *StoreSuite) TestMustCount() {
	s.NoError(s.store.Insert(ModelSchema, newModel("Joe", "", 1)))

	q := NewBaseQuery(ModelSchema)

	s.NotPanics(func() {
		s.Equal(int64(1), s.store.MustCount(q))
	})

	s.Panics(func() {
		s.errStore.MustCount(q)
	})
}

func (s *StoreSuite) TestTransaction() {
	err := s.store.Transaction(func(store *Store) error {
		s.NoError(store.Insert(ModelSchema, newModel("Joe", "", 1)))

		return store.Transaction(func(store *Store) error {
			return store.Insert(ModelSchema, newModel("Anna", "", 1))
		})
	})
	s.NoError(err)
	s.assertCount(2)
}

func (s *StoreSuite) TestTransaction_CantOpen() {
	err := s.errStore.Transaction(func(store *Store) error {
		return nil
	})
	s.Error(err)
}

func (s *StoreSuite) TestTransaction_Rollback() {
	err := s.store.Transaction(func(store *Store) error {
		s.NoError(store.Insert(ModelSchema, newModel("Joe", "", 1)))
		s.NoError(store.Insert(ModelSchema, newModel("Anna", "", 1)))
		return fmt.Errorf("kallax: we're never ever, ever, getting store together")
	})
	s.Error(err)
	s.assertCount(0)
}

func (s *StoreSuite) TestTransaction_RawExec() {
	err := s.store.Transaction(func(store *Store) error {
		_, err := store.RawExec("INSERT INTO model (name, email, age) VALUES ($1, $2, $3)", "foo", "bar", 1)
		s.NoError(err)
		return nil
	})
	s.NoError(err)
	s.assertCount(1)
}

func (s *StoreSuite) TestReload() {
	s.NoError(s.store.Insert(ModelSchema, newModel("Joe", "", 1)))

	// If we don't select all the fields, the records
	// retrieved will not be writable, as it could be a potential danger
	// to the user to save a partial model.
	q := NewBaseQuery(ModelSchema)
	q.Select(NewSchemaField("name"), ModelSchema.ID())
	rs, err := s.store.Find(q)
	s.NoError(err)
	s.True(rs.Next())

	var m = new(model)
	// First, we check that an empty model can't be reloaded, because it has
	// no ID
	s.Equal(ErrEmptyID, s.store.Reload(ModelSchema, m))
	record, err := rs.Get(ModelSchema)
	var ok bool
	m, ok = record.(*model)
	s.True(ok)
	s.NoError(err)

	// Model is not writable, as we said
	s.False(m.IsWritable())
	s.Equal(0, m.Age)

	_, err = s.store.Update(ModelSchema, m)
	s.Equal(ErrNotWritable, err)

	// Now, the model is reloaded with all the fields
	s.NoError(s.store.Reload(ModelSchema, m))

	// And so, it becomes writable
	s.True(m.IsWritable())
	s.Equal(1, m.Age)
}

func (s *StoreSuite) TestReload_Fail() {
	var m = newModel("Joe", "", 1)
	m.setPersisted()

	s.Error(s.errStore.Reload(ModelSchema, m))
}

func (s *StoreSuite) TestReload_NotFound() {
	var m = newModel("Joe", "", 1)
	m.ID = 1
	m.setPersisted()

	s.Equal(ErrNotFound, s.store.Reload(ModelSchema, m))
}

func (s *StoreSuite) TestFind_1to1() {
	m := newModel("Foo", "bar", 1)
	s.NoError(s.store.Insert(ModelSchema, m))

	rel := newRel(m.GetID(), "foo")
	s.NoError(s.store.Insert(RelSchema, rel))

	// just to see it does not randomly takes the most recent one
	s.NoError(s.store.Insert(RelSchema, newRel(new(NumericID), "foo")))

	q := NewBaseQuery(ModelSchema)
	q.AddRelation(RelSchema, "rel", OneToOne, nil)
	rs, err := s.store.Find(q)
	s.NoError(err)

	s.True(rs.Next())
	record, err := rs.Get(ModelSchema)
	s.NoError(err)
	model, ok := record.(*model)
	s.True(ok)

	s.Equal("Foo", model.Name)
	s.Equal("bar", model.Email)
	s.Equal(1, model.Age)
	s.NotNil(model.Rel)
	s.Equal(model.GetID(), model.Rel.VirtualColumn("model_id"))
	s.Equal("foo", model.Rel.Foo)
}

func (s *StoreSuite) rel1ToNFixtures() {
	m := newModel("Foo", "bar", 1)
	s.NoError(s.store.Insert(ModelSchema, m))

	rels := []string{"foo", "bar", "baz"}
	for _, v := range rels {
		rel := newRel(m.GetID(), v)
		s.NoError(s.store.Insert(RelSchema, rel))
	}

	s.NoError(s.store.Insert(RelSchema, newRel(new(NumericID), "qux")))
}

func (s *StoreSuite) TestFind_1toN() {
	s.rel1ToNFixtures()

	q := NewBaseQuery(ModelSchema)
	s.NoError(q.AddRelation(RelSchema, "rels", OneToMany, nil))
	rs, err := s.store.Find(q)
	s.NoError(err)

	s.True(rs.Next())
	record, err := rs.Get(ModelSchema)
	s.NoError(err)
	model, ok := record.(*model)
	s.True(ok)

	s.Equal("Foo", model.Name)
	s.Equal("bar", model.Email)
	s.Equal(1, model.Age)
	s.Nil(model.Rel)
	s.Require().Len(model.Rels, 3)
	s.Equal("foo", model.Rels[0].Foo)
	s.Equal("bar", model.Rels[1].Foo)
	s.Equal("baz", model.Rels[2].Foo)
}

func (s *StoreSuite) TestFind_1toN_Filter() {
	s.rel1ToNFixtures()

	q := NewBaseQuery(ModelSchema)
	s.NoError(q.AddRelation(RelSchema, "rels", OneToMany, Eq(NewSchemaField("foo"), "bar")))
	rs, err := s.store.Find(q)
	s.NoError(err)

	s.True(rs.Next())
	record, err := rs.Get(ModelSchema)
	s.NoError(err)
	model, ok := record.(*model)
	s.True(ok)

	s.Equal("Foo", model.Name)
	s.Equal("bar", model.Email)
	s.Equal(1, model.Age)
	s.Nil(model.Rel)
	s.Require().Len(model.Rels, 1)
	s.Equal("bar", model.Rels[0].Foo)
}

func (s *StoreSuite) TestFind_1toNAnd1to1() {
	s.rel1ToNFixtures()

	q := NewBaseQuery(ModelSchema)
	s.NoError(q.AddRelation(RelSchema, "rels", OneToMany, nil))
	s.NoError(q.AddRelation(RelSchema, "rel", OneToOne, nil))
	rs, err := s.store.Find(q)
	s.NoError(err)

	var foo string
	s.Equal(ErrRawScanBatching, rs.RawScan(&foo))

	s.True(rs.Next())
	record, err := rs.Get(ModelSchema)
	s.NoError(err)
	model, ok := record.(*model)
	s.True(ok)

	s.Equal("Foo", model.Name)
	s.Equal("bar", model.Email)
	s.Equal(1, model.Age)
	s.NotNil(model.Rel)
	s.Len(model.Rels, 3)
}

func (s *StoreSuite) TestFind_1toNMultiple() {
	rels := []string{"foo", "bar", "baz"}
	for i := 0; i < 100; i++ {
		m := newModel(fmt.Sprint(i), fmt.Sprint(i), i)
		s.NoError(s.store.Insert(ModelSchema, m))

		for _, v := range rels {
			s.NoError(s.store.Insert(RelSchema, newRel(m.GetID(), fmt.Sprintf("%s%d", v, i))))
		}
	}

	q := NewBaseQuery(ModelSchema)
	q.BatchSize(20)
	s.NoError(q.AddRelation(RelSchema, "rels", OneToMany, nil))
	rs, err := s.store.Find(q)
	s.NoError(err)

	var i int
	for rs.Next() {
		record, err := rs.Get(ModelSchema)
		s.NoError(err, "row #%d", i)

		model, ok := record.(*model)
		s.True(ok, "row #%d", i)

		s.Equal(fmt.Sprint(i), model.Name, "row #%d", i)
		s.Require().Len(model.Rels, 3, "row #%d", i)

		for j, v := range rels {
			s.Equal(model.GetID(), model.Rels[j].VirtualColumn("model_id"), "row #%d", i)
			s.Equal(fmt.Sprintf("%s%d", v, i), model.Rels[j].Foo, "row #%d", i)
		}
		i++
	}
	s.Equal(100, i)
}

func (s *StoreSuite) Fixtures_NtoM() {
	require := s.Require()

	lefts := []*throughLeft{
		newThroughLeft("a"),
		newThroughLeft("b"),
		newThroughLeft("c"),
	}

	for _, l := range lefts {
		require.NoError(s.store.Insert(ThroughLeftSchema, l))
	}

	rights := []*throughRight{
		newThroughRight("d"),
		newThroughRight("e"),
		newThroughRight("f"),
	}

	for _, r := range rights {
		require.NoError(s.store.Insert(ThroughRightSchema, r))
	}

	middles := []*throughMiddle{
		newThroughMiddle(lefts[0], rights[0]),
		newThroughMiddle(lefts[0], rights[1]),
		newThroughMiddle(lefts[1], rights[0]),
		newThroughMiddle(lefts[1], rights[1]),
		newThroughMiddle(lefts[2], rights[1]),
		newThroughMiddle(lefts[2], rights[2]),
	}

	for _, m := range middles {
		require.NoError(s.store.Insert(ThroughMiddleSchema, m))
	}
}

func (s *StoreSuite) TestFind_NtoM() {
	require := s.Require()
	s.Fixtures_NtoM()

	q := NewBaseQuery(ThroughLeftSchema)
	q.AddRelationThrough(ThroughRightSchema, ThroughMiddleSchema, "Rights", nil, nil)
	rs, err := s.store.Debug().Find(q)
	require.NoError(err)

	var results []*throughLeft
	for rs.Next() {
		r, err := rs.Get(ThroughLeftSchema)
		require.NoError(err)
		results = append(results, r.(*throughLeft))
	}

	require.Equal(3, len(results))

	var result = make(map[string][]string)
	for _, l := range results {
		var rights []string
		for _, r := range l.Rights {
			rights = append(rights, r.Name)
		}
		result[l.Name] = rights
	}

	require.Equal(map[string][]string{
		"a": {"d", "e"},
		"b": {"d", "e"},
		"c": {"e", "f"},
	}, result)
}

func (s *StoreSuite) TestFind_NtoM_Inverse() {
	require := s.Require()
	s.Fixtures_NtoM()

	q := NewBaseQuery(ThroughRightSchema)
	q.AddRelationThrough(ThroughLeftSchema, ThroughMiddleSchema, "Lefts", nil, nil)
	rs, err := s.store.Debug().Find(q)
	require.NoError(err)

	var results []*throughRight
	for rs.Next() {
		r, err := rs.Get(ThroughRightSchema)
		require.NoError(err)
		results = append(results, r.(*throughRight))
	}

	require.Equal(3, len(results))

	var result = make(map[string][]string)
	for _, r := range results {
		var lefts []string
		for _, r := range r.Lefts {
			lefts = append(lefts, r.Name)
		}
		result[r.Name] = lefts
	}

	require.Equal(map[string][]string{
		"d": {"a", "b"},
		"e": {"a", "b", "c"},
		"f": {"c"},
	}, result)
}

func (s *StoreSuite) assertModel(m *model) {
	var result model
	err := s.db.QueryRow("SELECT id, name, email, age FROM model WHERE id = $1", m.GetID()).
		Scan(result.GetID(), &result.Name, &result.Email, &result.Age)
	s.NoError(err)

	if err == nil {
		s.Equal(m.GetID(), result.GetID())
		s.Equal(m.Name, result.Name)
		s.Equal(m.Email, result.Email)
		s.Equal(m.Age, result.Age)
	}
}

func (s *StoreSuite) assertCount(n int64) {
	var count int64
	s.NoError(s.db.QueryRow("SELECT COUNT(*) FROM model").Scan(&count))
	s.Equal(n, count)
}

func (s *StoreSuite) assertNotExists(m *model) {
	var id int64
	err := s.db.QueryRow("SELECT id FROM model WHERE id = $1", m.GetID()).Scan(&id)
	s.Equal(sql.ErrNoRows, err, "record should not exist")
}

func TestStore(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

type store1 struct {
	*Store
}

func (s *store1) GenericStore() *Store {
	return s.Store
}

func (s *store1) SetGenericStore(store *Store) {
	s.Store = store
}

func TestStoreFrom(t *testing.T) {
	var s1 = &store1{new(Store)}
	var s2 store1

	require := require.New(t)
	require.NotEqual(s1.Store, s2.Store)

	StoreFrom(&s2, s1)
	require.Exactly(s1.Store, s2.Store)
}
