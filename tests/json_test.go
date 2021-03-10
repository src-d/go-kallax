package tests

import (
	"testing"

	kallax "github.com/loyalguru/go-kallax"
	"github.com/stretchr/testify/suite"
)

type JSONSuite struct {
	BaseTestSuite
}

func TestJSON(t *testing.T) {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS jsons (
			id uuid primary key,
			foo text,
			bar jsonb,
			baz jsonb,
			baz_slice jsonb
		)`,
	}
	suite.Run(t, &JSONSuite{NewBaseSuite(schema, "jsons")})
}

func (s *JSONSuite) TestSearchByField() {
	s.insertFixtures()
	q := NewJSONModelQuery().Where(
		kallax.Eq(Schema.JSONModel.Bar.Mux, "mux1"),
	)
	s.assertFound(q, "1")
}

func (s *JSONSuite) TestSearchByCustomField() {
	s.insertFixtures()
	q := NewJSONModelQuery().Where(
		kallax.Eq(kallax.AtJSONPath(Schema.JSONModel.Baz, kallax.JSONInt, "a", "0", "b"), 3),
	)

	s.assertFound(q, "2")

	q = NewJSONModelQuery().Where(
		kallax.Eq(kallax.AtJSONPath(Schema.JSONModel.Baz, kallax.JSONBool, "b"), true),
	)

	s.assertFound(q, "1")
}

func (s *JSONSuite) assertFound(q *JSONModelQuery, foos ...string) {
	require := s.Require()
	store := NewJSONModelStore(s.db)
	rs, err := store.Find(q)
	require.NoError(err)

	models, err := rs.All()
	require.NoError(err)
	require.Len(models, len(foos))
	for i, f := range foos {
		require.Equal(f, models[i].Foo)
	}
}

func (s *JSONSuite) insertFixtures() {
	store := NewJSONModelStore(s.db)

	m := NewJSONModel()
	m.Foo = "1"
	m.Baz = map[string]interface{}{
		"a": []interface{}{
			map[string]interface{}{
				"b": 1,
			},
			map[string]interface{}{
				"b": 2,
			},
		},
		"b": true,
	}
	m.Bar = &Bar{
		Qux: []Qux{
			{"schnooga1", 1, .5},
			{"schnooga2", 2, .6},
		},
		Mux: "mux1",
	}

	s.NoError(store.Insert(m))

	m = NewJSONModel()
	m.Foo = "2"
	m.Baz = map[string]interface{}{
		"a": []interface{}{
			map[string]interface{}{
				"b": 3,
			},
			map[string]interface{}{
				"b": 4,
			},
		},
		"b": false,
	}
	m.Bar = &Bar{
		Qux: []Qux{
			{"schnooga3", 3, .7},
			{"schnooga4", 4, .8},
		},
		Mux: "mux2",
	}

	s.NoError(store.Insert(m))
}
