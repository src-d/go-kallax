package tests

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	kallax "github.com/src-d/go-kallax"
	"github.com/stretchr/testify/suite"
)

type JSONSuite struct {
	BaseTestSuite
}

func TestJSON(t *testing.T) {
	schema := []string{
		`CREATE TABLE jsons (
			id uuid primary key,
			foo text,
			bar json,
			baz json
		)`,
	}
	suite.Run(t, &JSONSuite{
		BaseTestSuite{
			initQueries: schema,
		},
	})
}

func (s *JSONSuite) TestSearchByField() {
	s.insertFixtures()
	store := NewJSONModelStore(s.db)
	q := NewJSONModelQuery().Where(
		kallax.Eq(Schema.JSONModel.Bar.Mux, "mux1"),
	)

	rs, err := store.Find(q)
	s.NoError(err)

	models, err := rs.All()
	s.NoError(err)

	s.Len(models, 1)
	spew.Dump(models)
	s.Equal("foo", models[0].Foo)
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
				"c": 1,
			},
			map[string]interface{}{
				"c": 2,
			},
		},
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
