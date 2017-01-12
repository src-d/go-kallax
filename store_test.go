package storable

import (
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func (s *BaseSuite) TestStore_Insert(c *C) {
	p := NewPerson("foo")
	st := NewStore(s.db, "test")
	err := st.Insert(p)
	c.Assert(err, IsNil)
	c.Assert(p.IsNew(), Equals, false)

	r, err := st.Find(NewBaseQuery())
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 1)
	c.Assert(result[0].FirstName, Equals, "foo")
}

func (s *BaseSuite) TestStore_InsertOld(c *C) {
	p := NewPerson("foo")
	st := NewStore(s.db, "test")
	err := st.Insert(p)
	c.Assert(err, IsNil)

	err = st.Insert(p)
	c.Assert(err, Equals, ErrNonNewDocument)
}

func (s *BaseSuite) TestStore_Update(c *C) {
	p := NewPerson("foo")

	st := NewStore(s.db, "test")
	st.Insert(p)
	st.Insert(&Person{FirstName: "bar"})

	p.FirstName = "qux"
	err := st.Update(p)
	c.Assert(err, IsNil)

	q := NewBaseQuery()
	q.AddCriteria(bson.M{"firstname": "qux"})

	r, err := st.Find(q)
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 1)
	c.Assert(result[0].FirstName, Equals, "qux")
}

func (s *BaseSuite) TestStore_Save(c *C) {
	p := NewPerson("foo")
	p.SetId(bson.NewObjectId())

	st := NewStore(s.db, "test")
	updated, err := st.Save(p)
	c.Assert(err, IsNil)
	c.Assert(updated, Equals, false)
	c.Assert(p.IsNew(), Equals, false)

	p.FirstName = "qux"
	updated, err = st.Save(p)
	c.Assert(err, IsNil)
	c.Assert(updated, Equals, true)
	c.Assert(p.IsNew(), Equals, false)

	r, err := st.Find(NewBaseQuery())
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 1)
	c.Assert(result[0].FirstName, Equals, "qux")
}

func (s *BaseSuite) TestStore_UpdateNew(c *C) {
	p := NewPerson("foo")
	st := NewStore(s.db, "test")

	err := st.Update(p)
	c.Assert(err, Equals, ErrNewDocument)
}

func (s *BaseSuite) TestStore_Delete(c *C) {
	p := NewPerson("foo")
	st := NewStore(s.db, "test")
	st.Insert(p)

	err := st.Delete(p)
	c.Assert(err, IsNil)

	r, err := st.Find(NewBaseQuery())
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 0)
}

func (s *BaseSuite) TestStore_FindLimit(c *C) {
	st := NewStore(s.db, "test")
	st.Insert(NewPerson("foo"))
	st.Insert(NewPerson("bar"))

	q := NewBaseQuery()
	q.Limit(1)
	r, err := st.Find(q)
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 1)
	c.Assert(result[0].FirstName, Equals, "foo")
}

func (s *BaseSuite) TestStore_FindSkip(c *C) {
	st := NewStore(s.db, "test")
	st.Insert(NewPerson("foo"))
	st.Insert(NewPerson("bar"))

	q := NewBaseQuery()
	q.Skip(1)
	r, err := st.Find(q)
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 1)
	c.Assert(result[0].FirstName, Equals, "bar")
}

func (s *BaseSuite) TestStore_FindSort(c *C) {
	st := NewStore(s.db, "test")
	st.Insert(NewPerson("foo"))
	st.Insert(NewPerson("bar"))

	q := NewBaseQuery()
	q.Sort(Sort{{IdField, Desc}})
	r, err := st.Find(q)
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 2)
	c.Assert(result[0].FirstName, Equals, "bar")
	c.Assert(result[1].FirstName, Equals, "foo")
}

func (s *BaseSuite) TestStore_FindSelect(c *C) {
	p := NewPerson("foo")
	p.LastName = "qux"

	st := NewStore(s.db, "test")
	st.Insert(p)

	q := NewBaseQuery()
	q.Select(Select{{NewField("lastname", "string"), Exclude}})

	r, err := st.Find(q)
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 1)
	c.Assert(result[0].FirstName, Equals, "foo")
	c.Assert(result[0].LastName, Equals, "")
}

func (s *BaseSuite) TestStore_RawUpdate(c *C) {
	st := NewStore(s.db, "test")

	p1 := NewPerson("foo")
	p1.LastName = "bar"
	st.Insert(p1)

	p2 := NewPerson("bar")
	p2.LastName = "bar"
	st.Insert(p2)

	q := NewBaseQuery()
	q.AddCriteria(bson.M{"lastname": "bar"})

	err := st.RawUpdate(q, bson.M{"lastname": "qux"}, false)
	c.Assert(err, IsNil)

	q = NewBaseQuery()
	q.AddCriteria(bson.M{"lastname": "qux"})

	r, err := st.Find(q)
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 1)
	c.Assert(result[0].FirstName, Equals, "foo")
	c.Assert(result[0].LastName, Equals, "qux")
}

func (s *BaseSuite) TestStore_RawUpdateMulti(c *C) {
	st := NewStore(s.db, "test")

	p1 := NewPerson("foo")
	p1.LastName = "bar"
	st.Insert(p1)

	p2 := NewPerson("bar")
	p2.LastName = "bar"
	st.Insert(p2)

	q := NewBaseQuery()
	q.AddCriteria(bson.M{"lastname": "bar"})

	err := st.RawUpdate(q, bson.M{"lastname": "qux"}, true)
	c.Assert(err, IsNil)

	q = NewBaseQuery()
	q.AddCriteria(bson.M{"lastname": "qux"})

	r, err := st.Find(q)
	c.Assert(err, IsNil)

	var result []*Person
	c.Assert(r.All(&result), IsNil)
	c.Assert(result, HasLen, 2)
}

func (s *BaseSuite) TestStore_RawUpdateEmpty(c *C) {
	st := NewStore(s.db, "test")
	q := NewBaseQuery()
	err := st.RawUpdate(q, bson.M{"firstname": "qux"}, false)
	c.Assert(err, Equals, ErrEmptyQueryInRaw)
}

func (s *BaseSuite) TestStore_RawDelete(c *C) {
	st := NewStore(s.db, "test")
	st.Insert(NewPerson("bar"))
	st.Insert(NewPerson("bar"))

	q := NewBaseQuery()
	q.AddCriteria(bson.M{"firstname": "bar"})

	err := st.RawDelete(q, false)
	c.Assert(err, IsNil)

	q = NewBaseQuery()
	q.AddCriteria(bson.M{"firstname": "bar"})

	r, _ := st.Find(q)
	count, _ := r.Count()
	c.Assert(count, Equals, 1)
}

func (s *BaseSuite) TestStore_RawDeleteMulti(c *C) {
	st := NewStore(s.db, "test")
	st.Insert(NewPerson("bar"))
	st.Insert(NewPerson("bar"))

	q := NewBaseQuery()
	q.AddCriteria(bson.M{"firstname": "bar"})

	err := st.RawDelete(q, true)
	c.Assert(err, IsNil)

	q = NewBaseQuery()
	q.AddCriteria(bson.M{"firstname": "bar"})

	r, _ := st.Find(q)
	count, _ := r.Count()
	c.Assert(count, Equals, 0)
}

func (s *BaseSuite) TestStore_RawDeleteEmpty(c *C) {
	st := NewStore(s.db, "test")
	q := NewBaseQuery()
	err := st.RawDelete(q, false)
	c.Assert(err, Equals, ErrEmptyQueryInRaw)
}
