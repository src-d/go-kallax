package kallax

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type StoreSuite struct {
	suite.Suite
	db       *sql.DB
	store    *Store
	relStore *Store
}

func (s *StoreSuite) SetupTest() {
	var err error
	s.db, err = openTestDB()
	s.Nil(err)
	_, err = s.db.Exec(`CREATE TABLE model (
		id uuid PRIMARY KEY,
		name varchar(255) not null,
		email varchar(255) not null,
		age int not null
	)`)
	s.Nil(err)

	_, err = s.db.Exec(`CREATE TABLE rel (
		id uuid PRIMARY KEY,
		model_id uuid,
		foo text
	)`)
	s.Nil(err)

	s.store = NewStore(s.db, ModelSchema)
	s.relStore = NewStore(s.db, RelSchema)
}

func (s *StoreSuite) TearDownTest() {
	_, err := s.db.Exec("DROP TABLE model")
	s.Nil(err)
	_, err = s.db.Exec("DROP TABLE rel")
	s.Nil(err)
	s.Nil(s.db.Close())
}

func (s *StoreSuite) TestInsert() {
	m := newModel("a", "a@a.a", 1)
	s.Nil(s.store.Insert(m))
	s.True(m.IsPersisted(), "model should be persisted now")
	s.assertModel(m)
}

func (s *StoreSuite) TestInsert_NotNew() {
	var m model
	m.setPersisted()
	s.Equal(ErrNonNewDocument, s.store.Insert(&m))
}

func (s *StoreSuite) TestInsert_IDEmpty() {
	var m = new(model)
	s.Nil(s.store.Insert(m))
	s.False(m.ID.IsEmpty())
}

func (s *StoreSuite) TestUpdate() {
	var m = newModel("a", "a@a.a", 1)
	s.Nil(s.store.Insert(m))

	var newModel = newModel("a", "a@a.a", 1)
	newModel.SetID(m.ID)
	_, err := s.store.Update(newModel)
	s.Equal(ErrNewDocument, err)

	newModel.setPersisted()
	newModel.SetID(ID{})
	_, err = s.store.Update(newModel)
	s.Equal(ErrEmptyID, err)

	m.Age = 2
	m.Email = "b@b.b"
	m.Name = "b"
	rows, err := s.store.Update(m)
	s.Nil(err)
	s.Equal(int64(1), rows, "rows affected")
	s.assertModel(m)

	m.setWritable(false)
	_, err = s.store.Update(m)
	s.Equal(ErrNotWritable, err)
}

func (s *StoreSuite) TestSave() {
	m := newModel("a", "a@a.a", 1)
	updated, err := s.store.Save(m)
	s.Nil(err)
	s.False(updated)
	s.assertModel(m)

	m.Age = 5
	updated, err = s.store.Save(m)
	s.Nil(err)
	s.True(updated)

	m.setWritable(false)
	_, err = s.store.Save(m)
	s.Equal(ErrNotWritable, err)
}

func (s *StoreSuite) TestDelete() {
	m := newModel("a", "a@a.a", 1)
	s.Nil(s.store.Insert(m))
	s.assertModel(m)

	s.Nil(s.store.Delete(m))
	s.assertNotExists(m)
}

func (s *StoreSuite) TestRawQuery() {
	s.Nil(s.store.Insert(newModel("Joe", "", 1)))
	s.Nil(s.store.Insert(newModel("Jane", "", 2)))
	s.Nil(s.store.Insert(newModel("Anna", "", 2)))

	rs, err := s.store.RawQuery("SELECT name FROM model WHERE age > $1", 1)
	s.Nil(err)

	var names []string
	for rs.Next() {
		s.Equal(ErrRawScan, rs.Scan(nil))
		var name string
		s.Nil(rs.RawScan(&name))
		names = append(names, name)
	}
	s.Equal([]string{"Jane", "Anna"}, names)
}

func (s *StoreSuite) TestRawExec() {
	s.Nil(s.store.Insert(newModel("Joe", "", 1)))
	s.Nil(s.store.Insert(newModel("Jane", "", 2)))
	s.Nil(s.store.Insert(newModel("Anna", "", 2)))

	rows, err := s.store.RawExec("DELETE FROM model WHERE age > $1", 1)
	s.Nil(err)
	s.Equal(int64(2), rows)
}

func (s *StoreSuite) TestFind() {
	s.Nil(s.store.Insert(newModel("Joe", "", 1)))
	s.Nil(s.store.Insert(newModel("Jane", "", 2)))
	s.Nil(s.store.Insert(newModel("Anna", "", 2)))

	q := NewBaseQuery(ModelSchema)
	q.Select(f("name"))
	q.Where(Gt(f("age"), 1))

	rs := s.store.MustFind(q)

	var names []string
	for rs.Next() {
		var m = newModel("", "", 0)
		s.Nil(rs.Scan(m))
		s.True(m.IsPersisted())
		names = append(names, m.Name)
	}
	s.Equal([]string{"Jane", "Anna"}, names)
}

func (s *StoreSuite) TestCount() {
	s.Nil(s.store.Insert(newModel("Joe", "", 1)))
	s.Nil(s.store.Insert(newModel("Jane", "", 2)))
	s.Nil(s.store.Insert(newModel("Anna", "", 2)))

	q := NewBaseQuery(ModelSchema)
	q.Select(f("name"))
	q.Where(Gt(f("age"), 1))

	s.Equal(int64(2), s.store.MustCount(q))
}

func (s *StoreSuite) TestTransaction() {
	err := s.store.Transaction(func(store *Store) error {
		s.Nil(store.Insert(newModel("Joe", "", 1)))
		s.Nil(store.Insert(newModel("Anna", "", 1)))
		return nil
	})
	s.Nil(err)
	s.assertCount(2)
}

func (s *StoreSuite) TestTransactionRollback() {
	err := s.store.Transaction(func(store *Store) error {
		s.Nil(store.Insert(newModel("Joe", "", 1)))
		s.Nil(store.Insert(newModel("Anna", "", 1)))
		return fmt.Errorf("we're never ever, ever, getting store together")
	})
	s.NotNil(err)
	s.assertCount(0)
}

func (s *StoreSuite) TestReload() {
	s.Nil(s.store.Insert(newModel("Joe", "", 1)))

	// If we don't select all the fields, the records
	// retrieved will not be writable, as it could be a potential danger
	// to the user to save a partial model.
	q := NewBaseQuery(ModelSchema)
	q.Select(NewSchemaField("name"), ModelSchema.ID())
	rs, err := s.store.Find(q)
	s.Nil(err)
	s.True(rs.Next())

	var model model
	// First, we check that an empty model can't be reloaded, because it has
	// no ID
	s.Equal(ErrEmptyID, s.store.Reload(&model))
	s.Nil(rs.Scan(&model))

	// Model is not writable, as we said
	s.False(model.IsWritable())
	s.Equal(0, model.Age)

	_, err = s.store.Update(&model)
	s.Equal(ErrNotWritable, err)

	// Now, the model is reloaded with all the fields
	s.Nil(s.store.Reload(&model))

	// And so, it becomes writable
	s.True(model.IsWritable())
	s.Equal(1, model.Age)
}

func (s *StoreSuite) TestFind_1to1() {
	m := newModel("Foo", "bar", 1)
	s.Nil(s.store.Insert(m))

	rel := newRel(m.ID, "foo")
	s.Nil(s.relStore.Insert(rel))

	// just to see it does not randomly takes the most recent one
	s.Nil(s.relStore.Insert(newRel(NewID(), "foo")))

	q := NewBaseQuery(ModelSchema)
	q.AddRelation(RelSchema, "rel")
	rs, err := s.store.Find(q)
	s.Nil(err)

	var model model
	s.True(rs.Next())
	s.Nil(rs.Scan(&model))
	s.Equal("Foo", model.Name)
	s.Equal("bar", model.Email)
	s.Equal(1, model.Age)
	s.NotNil(model.Rel)
	s.Equal(model.ID, model.Rel.ModelID)
	s.Equal("foo", model.Rel.Foo)
}

func (s *StoreSuite) TestOperators() {
	cases := []struct {
		name  string
		cond  Condition
		count int64
	}{
		{"Eq", Eq(f("name"), "Joe"), 1},
		{"Gt", Gt(f("age"), 1), 2},
		{"Lt", Lt(f("age"), 2), 1},
		{"Neq", Neq(f("name"), "Joe"), 2},
		{"GtOrEq", GtOrEq(f("age"), 2), 2},
		{"LtOrEq", LtOrEq(f("age"), 3), 3},
		{"Not", Not(Eq(f("name"), "Joe")), 2},
		{"And", And(Neq(f("name"), "Joe"), Gt(f("age"), 1)), 2},
		{"Or", Or(Neq(f("name"), "Joe"), Eq(f("age"), 1)), 3},
		{"In", In(f("name"), "Joe", "Jane"), 2},
		{"NotIn", NotIn(f("name"), "Joe", "Jane"), 1},
	}

	s.Nil(s.store.Insert(newModel("Joe", "", 1)))
	s.Nil(s.store.Insert(newModel("Jane", "", 2)))
	s.Nil(s.store.Insert(newModel("Anna", "", 2)))

	for _, c := range cases {
		q := NewBaseQuery(ModelSchema)
		q.Where(c.cond)

		s.Equal(s.store.MustCount(q), c.count, c.name)
	}
}

func (s *StoreSuite) assertModel(m *model) {
	var result model
	err := s.db.QueryRow("SELECT id, name, email, age FROM model WHERE id = $1", m.ID).
		Scan(&result.ID, &result.Name, &result.Email, &result.Age)
	s.Nil(err)

	if err == nil {
		s.Equal(m.ID, result.ID)
		s.Equal(m.Name, result.Name)
		s.Equal(m.Email, result.Email)
		s.Equal(m.Age, result.Age)
	}
}

func (s *StoreSuite) assertCount(n int64) {
	var count int64
	s.Nil(s.db.QueryRow("SELECT COUNT(*) FROM model").Scan(&count))
	s.Equal(n, count)
}

func (s *StoreSuite) assertNotExists(m *model) {
	var id ID
	err := s.db.QueryRow("SELECT id FROM model WHERE id = $1", m.ID).Scan(&id)
	s.Equal(sql.ErrNoRows, err, "record should not exist")
}

func TestStore(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}
