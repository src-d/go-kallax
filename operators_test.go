package kallax

import (
	"database/sql"
	"testing"

	"github.com/loyalguru/go-kallax/types"
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
	s.store = NewStore(s.db)
}

func (s *OpsSuite) create(sql string) {
	_, err := s.db.Exec(sql)
	s.NoError(err)
}

func (s *OpsSuite) remove(table string) {
	_, err := s.db.Exec("DROP TABLE IF EXISTS " + table)
	s.NoError(err)
}

func (s *OpsSuite) TestOperators() {
	s.create(`CREATE TABLE model (
		id serial PRIMARY KEY,
		name varchar(255) not null,
		email varchar(255) not null,
		age int not null
	)`)
	defer s.remove("model")

	customGt := NewOperator(":col: > :arg:")
	customIn := NewMultiOperator(":col: IN :arg:")

	cases := []struct {
		name  string
		cond  Condition
		count int64
	}{
		{"Eq", Eq(f("name"), "Joe"), 1},
		{"Gt", Gt(f("age"), 1), 2},
		{"customGt", customGt(f("age"), 1), 2},
		{"Lt", Lt(f("age"), 2), 1},
		{"Neq", Neq(f("name"), "Joe"), 2},
		{"Like upper", Like(f("name"), "J%"), 2},
		{"Like lower", Like(f("name"), "j%"), 0},
		{"Ilike upper", Ilike(f("name"), "J%"), 2},
		{"Ilike lower", Ilike(f("name"), "j%"), 2},
		{"SimilarTo", SimilarTo(f("name"), "An{2}a"), 1},
		{"NotSimilarTo", NotSimilarTo(f("name"), "An{2}a"), 2},
		{"GtOrEq", GtOrEq(f("age"), 2), 2},
		{"LtOrEq", LtOrEq(f("age"), 3), 3},
		{"Not", Not(Eq(f("name"), "Joe")), 2},
		{"And", And(Neq(f("name"), "Joe"), Gt(f("age"), 1)), 2},
		{"Or", Or(Neq(f("name"), "Joe"), Eq(f("age"), 1)), 3},
		{"In", In(f("name"), "Joe", "Jane"), 2},
		{"customIn", customIn(f("name"), "Joe", "Jane"), 2},
		{"NotIn", NotIn(f("name"), "Joe", "Jane"), 1},
		{"MatchRegexCase upper", MatchRegexCase(f("name"), "J.*"), 2},
		{"MatchRegexCase lower", MatchRegexCase(f("name"), "j.*"), 0},
		{"MatchRegex upper", MatchRegex(f("name"), "J.*"), 2},
		{"MatchRegex lower", MatchRegex(f("name"), "j.*"), 2},
		{"NotMatchRegexCase upper", NotMatchRegexCase(f("name"), "J.*"), 1},
		{"NotMatchRegexCase lower", NotMatchRegexCase(f("name"), "j.*"), 3},
		{"NotMatchRegex upper", NotMatchRegex(f("name"), "J.*"), 1},
		{"NotMatchRegex lower", NotMatchRegex(f("name"), "j.*"), 1},
	}

	s.Nil(s.store.Insert(ModelSchema, newModel("Joe", "", 1)))
	s.Nil(s.store.Insert(ModelSchema, newModel("Jane", "", 2)))
	s.Nil(s.store.Insert(ModelSchema, newModel("Anna", "", 2)))

	for _, c := range cases {
		q := NewBaseQuery(ModelSchema)
		q.Where(c.cond)

		s.Equal(c.count, s.store.Debug().MustCount(q), c.name)
	}
}

func (s *OpsSuite) TestArrayOperators() {
	s.create(`CREATE TABLE slices (
		id uuid PRIMARY KEY,
		elems bigint[]
	)`)
	defer s.remove("slices")

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

	_, err := s.db.Exec("INSERT INTO slices (id,elems) VALUES ($1, $2)", NewULID(), types.Slice([]int64{1, 2, 3}))
	s.NoError(err)

	for _, c := range cases {
		q := NewBaseQuery(SlicesSchema)
		q.Where(c.cond)
		cnt, err := s.store.Count(q)
		s.NoError(err, c.name)
		s.Equal(c.ok, cnt > 0, "success: %s", c.name)
	}
}

type object map[string]interface{}

type array []interface{}

func (s *OpsSuite) TestJSONOperators() {
	s.create(`CREATE TABLE jsons (
		id uuid primary key,
		elem jsonb
	)`)
	defer s.remove("jsons")

	f := f("elem")
	cases := []struct {
		name string
		cond Condition
		n    int64
	}{
		{"JSONIsObject", JSONIsObject(f), 2},
		{"JSONIsArray", JSONIsArray(f), 3},
		{"JSONContains", JSONContains(f, object{"a": 1}), 1},
		{"JSONContainedBy", JSONContainedBy(f, object{
			"a": 1,
			"b": 2,
			"c": 3,
			"d": 1,
		}), 1},
		{"JSONContainsAnyKey with array match", JSONContainsAnyKey(f, "a", "c"), 3},
		{"JSONContainsAnyKey", JSONContainsAnyKey(f, "b", "e"), 2},
		{"JSONContainsAllKeys with array match", JSONContainsAllKeys(f, "a", "c"), 3},
		{"JSONContainsAllKeys", JSONContainsAllKeys(f, "b", "e"), 0},
		{"JSONContainsAllKeys only objects", JSONContainsAllKeys(f, "a", "b", "c"), 2},
		{"JSONContainsAny", JSONContainsAny(f,
			object{"a": 1},
			object{"a": true},
		), 2},
	}

	var records = []interface{}{
		array{"a", "c", "d"},
		object{
			"a": true,
			"b": array{1, 2, 3},
			"c": object{"d": "foo"},
		},
		object{
			"a": 1,
			"b": 2,
			"c": 3,
		},
		array{.5, 1., 1.5},
		array{1, 2, 3},
	}

	for _, r := range records {
		_, err := s.db.Exec("INSERT INTO jsons (id,elem) VALUES ($1, $2)", NewULID(), types.JSON(r))
		s.NoError(err)
	}

	for _, c := range cases {
		q := NewBaseQuery(JsonsSchema)
		q.Where(c.cond)
		cnt, err := s.store.Count(q)
		s.NoError(err, c.name)
		s.Equal(c.n, cnt, "should retrieve %d records: %s", c.n, c.name)
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

var JsonsSchema = &BaseSchema{
	alias: "_js",
	table: "jsons",
	id:    f("id"),
	columns: []SchemaField{
		f("id"),
		f("elem"),
	},
}
