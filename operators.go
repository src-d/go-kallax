package kallax

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

// Condition represents a condition of filtering in a query.
type Condition interface {
	squirrel.Sqlizer
	isCondition()
}

type (
	eq struct {
		squirrel.Eq
	}
	lt struct {
		squirrel.Lt
	}
	gt struct {
		squirrel.Gt
	}
	ltOrEq struct {
		squirrel.LtOrEq
	}
	gtOrEq struct {
		squirrel.GtOrEq
	}
	neq struct {
		squirrel.NotEq
	}
	or struct {
		squirrel.Or
	}
	and struct {
		squirrel.And
	}
	not struct {
		cond Condition
	}
)

// Eq returns a condition that will be true when `col` is equal to `value`.
func Eq(col string, value interface{}) Condition {
	return eq{squirrel.Eq{col: value}}
}

// Lt returns a condition that will be true when `col` is lower than `value`.
func Lt(col string, value interface{}) Condition {
	return lt{squirrel.Lt{col: value}}
}

// Gt returns a condition that will be true when `col` is greater than `value`.
func Gt(col string, n interface{}) Condition {
	return gt{squirrel.Gt{col: n}}
}

// LtOrEq returns a condition that will be true when `col` is lower than
// `value` or equal.
func LtOrEq(col string, n interface{}) Condition {
	return ltOrEq{squirrel.LtOrEq{col: n}}
}

// GtOrEq returns a condition that will be true when `col` is greater than
// `value` or equal.
func GtOrEq(col string, n interface{}) Condition {
	return gtOrEq{squirrel.GtOrEq{col: n}}
}

// Neq returns a condition that will be true when `col` is not `value`.
func Neq(col string, n interface{}) Condition {
	return neq{squirrel.NotEq{col: n}}
}

// Or returns the given conditions joined by logical ors.
func Or(conds ...Condition) Condition {
	return or{squirrel.Or(condsToSqlizers(conds))}
}

// And returns the given conditions joined by logical ands.
func And(conds ...Condition) Condition {
	return and{squirrel.And(condsToSqlizers(conds))}
}

// Not returns the given condition negated.
func Not(cond Condition) Condition {
	return not{cond}
}

// Lt returns a condition that will be true when `col` is lower than `value`.
func condsToSqlizers(conds []Condition) []squirrel.Sqlizer {
	var result = make([]squirrel.Sqlizer, len(conds))
	for i, v := range conds {
		result[i] = v
	}
	return result
}

func (n not) ToSql() (string, []interface{}, error) {
	sql, args, err := n.cond.ToSql()
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("NOT %s", sql), args, err
}

func (eq) isCondition()     {}
func (lt) isCondition()     {}
func (gt) isCondition()     {}
func (gtOrEq) isCondition() {}
func (ltOrEq) isCondition() {}
func (neq) isCondition()    {}
func (or) isCondition()     {}
func (and) isCondition()    {}
func (not) isCondition()    {}
