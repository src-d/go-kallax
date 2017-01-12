package operators

import (
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func (s *OperatorsSuite) TestExists(c *check.C) {
	exists := Exists(Foo, true)
	c.Assert(exists, check.DeepEquals, bson.M{"foo": bson.M{"$exists": true}})
}

func (s *OperatorsSuite) TestType(c *check.C) {
	t := Type(Foo, Double)
	c.Assert(t, check.DeepEquals, bson.M{"foo": bson.M{"$type": Double}})
}
