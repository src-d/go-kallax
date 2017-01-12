package operators

import (
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func (s *OperatorsSuite) TestMod(c *check.C) {
	mod := Mod(Foo, 42, 82)
	c.Assert(mod, check.DeepEquals, bson.M{"foo": bson.M{"$mod": []float64{42, 82}}})
}

func (s *OperatorsSuite) TestRegEx(c *check.C) {
	re := RegEx(Foo, ".*", "i")
	c.Assert(re, check.DeepEquals, bson.M{
		"foo": bson.M{"$regex": bson.RegEx{Pattern: ".*", Options: "i"}},
	})
}

func (s *OperatorsSuite) TestText(c *check.C) {
	text := Text(Foo, "foo", "none")
	c.Assert(text, check.DeepEquals, bson.M{
		"foo": bson.M{"$text": bson.M{"$search": "foo", "$language": "none"}},
	})
}

func (s *OperatorsSuite) TestWhere(c *check.C) {
	where := Where(Foo, "foo", nil)
	c.Assert(where, check.DeepEquals, bson.M{
		"foo": bson.M{"$where": bson.JavaScript{Code: "foo", Scope: interface{}(nil)}},
	})
}
