package operators

import (
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func (s *OperatorsSuite) TestEq(c *check.C) {
	eq := Eq(Foo, "bar")
	c.Assert(eq, check.DeepEquals, bson.M{"foo": bson.M{"$eq": "bar"}})
}

func (s *OperatorsSuite) TestGt(c *check.C) {
	gt := Gt(Foo, "bar")
	c.Assert(gt, check.DeepEquals, bson.M{"foo": bson.M{"$gt": "bar"}})
}

func (s *OperatorsSuite) TestGte(c *check.C) {
	gte := Gte(Foo, "bar")
	c.Assert(gte, check.DeepEquals, bson.M{"foo": bson.M{"$gte": "bar"}})
}

func (s *OperatorsSuite) TestLt(c *check.C) {
	lt := Lt(Foo, "bar")
	c.Assert(lt, check.DeepEquals, bson.M{"foo": bson.M{"$lt": "bar"}})
}

func (s *OperatorsSuite) TestLte(c *check.C) {
	lte := Lte(Foo, "bar")
	c.Assert(lte, check.DeepEquals, bson.M{"foo": bson.M{"$lte": "bar"}})
}

func (s *OperatorsSuite) TestNe(c *check.C) {
	ne := Ne(Foo, "bar")
	c.Assert(ne, check.DeepEquals, bson.M{"foo": bson.M{"$ne": "bar"}})
}

func (s *OperatorsSuite) TestIn(c *check.C) {
	in := In(Foo, "bar", "qux")
	c.Assert(in, check.DeepEquals, bson.M{
		"foo": bson.M{"$in": []interface{}{"bar", "qux"}},
	})
}

func (s *OperatorsSuite) TestNin(c *check.C) {
	nin := Nin(Foo, "bar", "qux")
	c.Assert(nin, check.DeepEquals, bson.M{
		"foo": bson.M{"$nin": []interface{}{"bar", "qux"}},
	})
}
