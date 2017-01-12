package operators

import (
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func (s *OperatorsSuite) TestOr(c *check.C) {
	or := Or(bson.M{"foo": "qux"}, bson.M{"qux": "qux"})
	c.Assert(or, check.DeepEquals, bson.M{
		"$or": []bson.M{
			bson.M{"foo": "qux"},
			bson.M{"qux": "qux"},
		},
	})
}

func (s *OperatorsSuite) TestAnd(c *check.C) {
	and := And(bson.M{"foo": "qux"}, bson.M{"qux": "qux"})
	c.Assert(and, check.DeepEquals, bson.M{
		"$and": []bson.M{
			bson.M{"foo": "qux"},
			bson.M{"qux": "qux"},
		},
	})
}

func (s *OperatorsSuite) TestNot(c *check.C) {
	not := Not(bson.M{"foo": "qux"})
	c.Assert(not, check.DeepEquals, bson.M{"foo": bson.M{"$not": "qux"}})
}

func (s *OperatorsSuite) TestNor(c *check.C) {
	nor := Nor(bson.M{"foo": "qux"}, bson.M{"qux": "qux"})
	c.Assert(nor, check.DeepEquals, bson.M{
		"$nor": []bson.M{
			bson.M{"foo": "qux"},
			bson.M{"qux": "qux"},
		},
	})
}
