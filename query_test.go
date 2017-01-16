package kallax

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/suite"
)

func TestBaseQuery(t *testing.T) {
	suite.Run(t, new(QuerySuite))
}

type QuerySuite struct {
	suite.Suite
	q *BaseQuery
}

func (s *QuerySuite) SetupTest() {
	s.q = NewBaseQuery("foo")
}

func (s *QuerySuite) TestSelect() {
	s.q.Select("a", "b", "c")
	s.Equal(columnSet{"a", "b", "c"}, s.q.columns)
}

func (s *QuerySuite) TestSelectNot() {
	s.q.SelectNot("a", "b", "c")
	s.Equal(columnSet{"a", "b", "c"}, s.q.excludedColumns)
}

func (s *QuerySuite) TestSelectNotSelectSelectNot() {
	s.q.SelectNot("a", "b")
	s.q.Select("a", "c")
	s.q.SelectNot("a")
	s.Equal([]string{"c"}, s.q.selectedColumns())
}

func (s *QuerySuite) TestSelectSelectNot() {
	s.q.Select("a", "c")
	s.q.SelectNot("a")
	s.Equal([]string{"c"}, s.q.selectedColumns())
}

func (s *QuerySuite) TestCopy() {
	s.q.Select("a", "b", "c")
	s.q.SelectNot("foo")
	s.q.BatchSize(30)
	s.q.Limit(2)
	s.q.Offset(30)
	copy := s.q.Copy()

	s.Equal(s.q, copy)
	s.NotEqual(unsafe.Pointer(s.q), unsafe.Pointer(copy))
}

func (s *QuerySuite) TestSelectedColumns() {
	s.q.Select("a", "b", "c")
	s.q.SelectNot("b")
	s.Equal([]string{"a", "c"}, s.q.selectedColumns())
}

func (s *QuerySuite) TestOrder() {
	s.q.Select("foo")
	s.q.Order(Asc("bar"))
	s.q.Order(Desc("baz"))

	s.assertSql("SELECT foo FROM foo ORDER BY bar ASC, baz DESC")
}

func (s *QuerySuite) TestWhere() {
	s.q.Select("foo")
	s.q.Where(Eq("foo", 5))
	s.q.Where(Eq("bar", "baz"))

	s.assertSql("SELECT foo FROM foo WHERE foo = $1 AND bar = $2")
}

func (s *QuerySuite) TestString() {
	s.q.Select("foo")
	s.Equal("SELECT foo FROM foo", s.q.String())
}

func (s *QuerySuite) assertSql(sql string) {
	_, builder := s.q.compile()
	result, _, err := builder.ToSql()
	s.Nil(err)
	s.Equal(sql, result)
}
