package storable

import (
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func (s *BaseSuite) TestBaseQuery_AddCriteria(c *C) {
	q := NewBaseQuery()
	q.AddCriteria(bson.M{"foo": "foo"})
	q.AddCriteria(bson.M{"qux": "qux"})

	c.Assert(q.GetCriteria(), DeepEquals, bson.M{
		"$and": []bson.M{
			bson.M{"foo": "foo"},
			bson.M{"qux": "qux"},
		},
	})
}

func (s *BaseSuite) TestBaseQuery_String(c *C) {
	q := NewBaseQuery()
	q.AddCriteria(bson.M{"foo": "foo"})
	q.AddCriteria(bson.M{"qux": "qux"})

	c.Assert(q.String(), Equals, `{"$and":[{"foo":"foo"},{"qux":"qux"}]}`)
}
