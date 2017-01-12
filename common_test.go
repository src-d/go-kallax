package kallax

import (
	"testing"

	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	testMongoHost = "127.0.0.1:27017"
)

func Test(t *testing.T) { TestingT(t) }

type BaseSuite struct {
	db *mgo.Database
}

var _ = Suite(&BaseSuite{})

func (s *BaseSuite) SetUpTest(c *C) {
	conn, err := mgo.Dial(testMongoHost)
	if err != nil {
		panic(err)
	}
	s.db = conn.DB(bson.NewObjectId().Hex())
}

func (s *BaseSuite) TestMap_Key(c *C) {
	m := NewMap("foo."+mapPlaceholder, "string")
	f := m.Key("qux")
	c.Assert(f.String(), Equals, "foo.qux")
	c.Assert(f.Type(), Equals, "string")
}

func (s *BaseSuite) TestSort_ToList(c *C) {
	sort := Sort{{NewField("foo", ""), Asc}}
	c.Assert(sort.ToList(), DeepEquals, []string{"foo"})

	sort = Sort{{NewField("foo", ""), Desc}}
	c.Assert(sort.ToList(), DeepEquals, []string{"-foo"})

	sort = Sort{{NewField("foo", ""), Asc}, {NewField("qux", ""), Desc}}
	c.Assert(sort.ToList(), DeepEquals, []string{"foo", "-qux"})
}

func (s *BaseSuite) TestSort_IsEmpty(c *C) {
	sort := Sort{{NewField("foo", ""), Asc}}
	c.Assert(sort.IsEmpty(), Equals, false)

	sort = Sort{}
	c.Assert(sort.IsEmpty(), Equals, true)
}

func (s *BaseSuite) TestSelect_ToMap(c *C) {
	sel := Select{{NewField("foo", ""), Exclude}}
	c.Assert(sel.ToMap(), DeepEquals, bson.M{"foo": 0})

	sel = Select{{NewField("foo", ""), Include}}
	c.Assert(sel.ToMap(), DeepEquals, bson.M{"foo": 1})
}

func (s *BaseSuite) TestSelect_IsEmpty(c *C) {
	sel := Select{{NewField("foo", ""), Include}}
	c.Assert(sel.IsEmpty(), Equals, false)

	sel = Select{}
	c.Assert(sel.IsEmpty(), Equals, true)
}

func (s *BaseSuite) TearDownTest(c *C) {
	s.db.DropDatabase()
}

type Person struct {
	Document  `bson:",inline"`
	FirstName string
	LastName  string
	Gender    string
}

func NewPerson(name string) *Person {
	doc := &Person{FirstName: name}
	doc.SetIsNew(true)

	return doc
}
