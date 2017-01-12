package storable

import (
	"encoding/json"

	"gopkg.in/src-d/storable.v1/operators"

	"gopkg.in/mgo.v2/bson"
)

type Query interface {
	GetCriteria() bson.M
	Sort(s Sort)
	Limit(l int)
	Skip(s int)
	Select(p Select)
	GetSort() Sort
	GetLimit() int
	GetSkip() int
	GetSelect() Select
}

type BaseQuery struct {
	clauses     []bson.M
	limit, skip int
	sort        Sort
	selector    Select
}

func NewBaseQuery() *BaseQuery {
	return &BaseQuery{clauses: make([]bson.M, 0)}
}

// AddCriteria adds a new mathing expression to the query, all the expressions
// are merged on a $and expression.
//
// Use operators package instead of build expresion by hand:
//
//  import . "gopkg.in/src-d/storable.v1/operators"
//
//  func (q *YourQuery) FindNonZeroRecords() {
//      // All the Fields are defined on the Schema generated variable
//      size := NewField("size", "int")
//      q.AddCriteria(Gt(size, 0))
//  }
func (q *BaseQuery) AddCriteria(expr bson.M) {
	q.clauses = append(q.clauses, expr)
}

// GetCriteria returns a valid bson.M used internally by Store.
func (q *BaseQuery) GetCriteria() bson.M {
	if len(q.clauses) == 0 {
		return nil
	}

	return operators.And(q.clauses...)
}

// Sort sets the sorting cristeria of the query.
func (q *BaseQuery) Sort(s Sort) {
	q.sort = s
}

// Limit sets the limit of the query.
func (q *BaseQuery) Limit(l int) {
	q.limit = l
}

// Skip sets the skip of the query.
func (q *BaseQuery) Skip(s int) {
	q.skip = s
}

// Select specifies the fields to return using projection operators.
// http://docs.mongodb.org/manual/reference/operator/projection/
func (q *BaseQuery) Select(projection Select) {
	q.selector = projection
}

// GetSort return the current sorting preferences of the query.
func (q *BaseQuery) GetSort() Sort {
	return q.sort
}

// GetLimit return the current limit preferences of the query.
func (q *BaseQuery) GetLimit() int {
	return q.limit
}

// GetSkip return the current skip preferences of the query.
func (q *BaseQuery) GetSkip() int {
	return q.skip
}

// GetSelect return the current select preferences of the query.
func (q *BaseQuery) GetSelect() Select {
	return q.selector
}

// Strings return a json representation of the criteria. Sorry but this is not
// fully compatible with the MongoDb CLI.
func (q *BaseQuery) String() string {
	j, _ := json.Marshal(q.GetCriteria())

	return string(j)
}
