package operators

import (
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func (s *OperatorsSuite) TestAll(c *check.C) {
	all := All(Foo, "qux", "bar")
	c.Assert(all, check.DeepEquals, bson.M{"foo": bson.M{"$all": []interface{}{"qux", "bar"}}})
}

func (s *OperatorsSuite) TestSize(c *check.C) {
	size := Size(Foo, 2)
	c.Assert(size, check.DeepEquals, bson.M{"foo": bson.M{"$size": 2}})
}
