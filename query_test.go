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
	s.q = NewBaseQuery(ModelSchema)
}

func (s *QuerySuite) TestSelect() {
	s.q.Select(f("a"), f("b"), f("c"))
	s.Equal(columnSet{f("a"), f("b"), f("c")}, s.q.columns)
}

func (s *QuerySuite) TestSelectNot() {
	s.q.SelectNot(f("a"), f("b"), f("c"))
	s.Equal(columnSet{f("a"), f("b"), f("c")}, s.q.excludedColumns)
}

func (s *QuerySuite) TestSelectNotSelectSelectNot() {
	s.q.SelectNot(f("a"), f("b"))
	s.q.Select(f("a"), f("c"))
	s.q.SelectNot(f("a"))
	s.Equal([]SchemaField{f("c")}, s.q.selectedColumns())
}

func (s *QuerySuite) TestSelectSelectNot() {
	s.q.Select(f("a"), f("c"))
	s.q.SelectNot(f("a"))
	s.Equal([]SchemaField{f("c")}, s.q.selectedColumns())
}

func (s *QuerySuite) TestCopy() {
	s.q.Select(f("a"), f("b"), f("c"))
	s.q.SelectNot(f("foo"))
	s.q.BatchSize(30)
	s.q.Limit(2)
	s.q.Offset(30)
	copy := s.q.Copy()

	s.Equal(s.q, copy)
	s.NotEqual(unsafe.Pointer(s.q), unsafe.Pointer(copy))
}

func (s *QuerySuite) TestSelectedColumns() {
	s.q.Select(f("a"), f("b"), f("c"))
	s.q.SelectNot(f("b"))
	s.Equal([]SchemaField{f("a"), f("c")}, s.q.selectedColumns())
}

func (s *QuerySuite) TestOrder() {
	s.q.Select(f("foo"))
	s.q.Order(Asc(f("bar")))
	s.q.Order(Desc(f("baz")))

	s.assertSql("SELECT __model.foo FROM model __model ORDER BY __model.bar ASC, __model.baz DESC")
}

func (s *QuerySuite) TestWhere() {
	s.q.Select(f("foo"))
	s.q.Where(Eq(f("foo"), 5))
	s.q.Where(Eq(f("bar"), "baz"))

	s.assertSql("SELECT __model.foo FROM model __model WHERE __model.foo = $1 AND __model.bar = $2")
}

func (s *QuerySuite) TestString() {
	s.q.Select(f("foo"))
	s.Equal("SELECT __model.foo FROM model __model", s.q.String())
}

func (s *QuerySuite) TestAddRelation() {
	s.Nil(s.q.AddRelation(RelSchema, "rel", OneToOne, nil))
	s.Equal("SELECT __model.id, __model.name, __model.email, __model.age, __rel_rel.id, __rel_rel.model_id, __rel_rel.foo FROM model __model LEFT JOIN rel __rel_rel ON (__rel_rel.model_id = __model.id)", s.q.String())
}

func (s *QuerySuite) TestAddRelation_Inverse() {
	s.Nil(s.q.AddRelation(RelSchema, "rel_inv", OneToOne, nil))
	s.Equal("SELECT __model.id, __model.name, __model.email, __model.age, __rel_rel_inv.id, __rel_rel_inv.model_id, __rel_rel_inv.foo FROM model __model LEFT JOIN rel __rel_rel_inv ON (__rel_rel_inv.id = __model.model_id)", s.q.String())
}

func (s *QuerySuite) TestAddRelation_ManyToMany() {
	err := s.q.AddRelation(RelSchema, "rel", ManyToMany, nil)
	s.Equal(ErrManyToManyNotSupported, err)
}

func (s *QuerySuite) TestAddRelation_FKNotFound() {
	s.Error(s.q.AddRelation(RelSchema, "fooo", OneToOne, nil))
}

func (s *QuerySuite) assertSql(sql string) {
	_, builder := s.q.compile()
	result, _, err := builder.ToSql()
	s.Nil(err)
	s.Equal(sql, result)
}
