package kallax

import (
	"database/sql"
	"testing"

	"github.com/src-d/go-kallax/types"
	"github.com/stretchr/testify/suite"
)

type OpsSuite struct {
	suite.Suite
	db    *sql.DB
	store *Store
}

func (s *OpsSuite) SetupTest() {
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

	_, err = s.db.Exec(`CREATE TABLE slices (
		id uuid PRIMARY KEY,
		elems bigint[]
	)`)
	s.Nil(err)

	s.store = NewStore(s.db)
}

func (s *OpsSuite) TearDownTest() {
	_, err := s.db.Exec("DROP TABLE slices")
	s.NoError(err)

	_, err = s.db.Exec("DROP TABLE model")
	s.NoError(err)
}

func (s *OpsSuite) TestOperators() {
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

	s.Nil(s.store.Insert(ModelSchema, newModel("Joe", "", 1)))
	s.Nil(s.store.Insert(ModelSchema, newModel("Jane", "", 2)))
	s.Nil(s.store.Insert(ModelSchema, newModel("Anna", "", 2)))

	for _, c := range cases {
		q := NewBaseQuery(ModelSchema)
		q.Where(c.cond)

		s.Equal(s.store.MustCount(q), c.count, c.name)
	}
}

func (s *OpsSuite) TestArrayOperators() {
	f := f("elems")

	cases := []struct {
		name string
		cond Condition
		ok   bool
	}{
		{"ArrayEq", ArrayEq(f, 1, 2, 3), true},
		{"ArrayEq fail", ArrayEq(f, 1, 2, 2), false},
		{"ArrayNotEq", ArrayNotEq(f, 1, 2, 2), true},
		{"ArrayNotEq fail", ArrayNotEq(f, 1, 2, 3), false},
		{"ArrayGt", ArrayGt(f, 1, 2, 2), true},
		{"ArrayGt all eq", ArrayGt(f, 1, 2, 3), false},
		{"ArrayGt some lt", ArrayGt(f, 1, 3, 1), false},
		{"ArrayLt", ArrayLt(f, 1, 2, 4), true},
		{"ArrayLt all eq", ArrayLt(f, 1, 2, 3), false},
		{"ArrayLt some gt", ArrayLt(f, 1, 1, 4), false},
		{"ArrayGtOrEq", ArrayGtOrEq(f, 1, 2, 2), true},
		{"ArrayGtOrEq all eq", ArrayGtOrEq(f, 1, 2, 3), true},
		{"ArrayGtOrEq some lt", ArrayGtOrEq(f, 1, 3, 1), false},
		{"ArrayLtOrEq", ArrayLtOrEq(f, 1, 2, 4), true},
		{"ArrayLtOrEq all eq", ArrayLtOrEq(f, 1, 2, 3), true},
		{"ArrayLtOrEq some gt", ArrayLtOrEq(f, 1, 1, 4), false},
		{"ArrayContains", ArrayContains(f, 1, 2), true},
		{"ArrayContains fail", ArrayContains(f, 5, 6), false},
		{"ArrayContainedBy", ArrayContainedBy(f, 1, 2, 3, 5, 6), true},
		{"ArrayContainedBy fail", ArrayContainedBy(f, 1, 2, 5, 6), false},
		{"ArrayOverlap", ArrayOverlap(f, 5, 1, 7), true},
		{"ArrayOverlap fail", ArrayOverlap(f, 6, 7, 8, 9), false},
	}

	_, err := s.db.Exec("INSERT INTO slices (id,elems) VALUES ($1, $2)", NewID(), types.Slice([]int64{1, 2, 3}))
	s.NoError(err)

	for _, c := range cases {
		q := NewBaseQuery(SlicesSchema)
		q.Where(c.cond)
		cnt, err := s.store.Count(q)
		s.NoError(err, c.name)
		s.Equal(c.ok, cnt > 0, "success: %s", c.name)
	}
}

func TestOperators(t *testing.T) {
	suite.Run(t, new(OpsSuite))
}

var SlicesSchema = &BaseSchema{
	alias: "_sl",
	table: "slices",
	id:    f("id"),
	columns: []SchemaField{
		f("id"),
		f("elems"),
	},
}
