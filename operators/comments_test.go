package operators

import (
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func (s *OperatorsSuite) TestComment(c *check.C) {
	comment := Comment("foo")
	c.Assert(comment, check.DeepEquals, bson.M{"$comment": "foo"})
}
