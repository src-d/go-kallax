package kallax

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/src-d/go-kallax/types"

	"github.com/Masterminds/squirrel"
)

// ScalarCond returns a kallax.Condition that compares a property with the passed
// values, considering its scalar values (eq, gt, gte, lt, lte, neq)
type ScalarCond func(col SchemaField, value interface{}) Condition

// Condition represents a condition of filtering in a query.
type Condition func(Schema) squirrel.Sqlizer

// Eq returns a condition that will be true when `col` is equal to `value`.
func Eq(col SchemaField, value interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.Eq{col.QualifiedName(schema): value}
	}
}

// Lt returns a condition that will be true when `col` is lower than `value`.
func Lt(col SchemaField, value interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.Lt{col.QualifiedName(schema): value}
	}
}

// Gt returns a condition that will be true when `col` is greater than `value`.
func Gt(col SchemaField, value interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.Gt{col.QualifiedName(schema): value}
	}
}

// LtOrEq returns a condition that will be true when `col` is lower than
// `value` or equal.
func LtOrEq(col SchemaField, value interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.LtOrEq{col.QualifiedName(schema): value}
	}
}

// GtOrEq returns a condition that will be true when `col` is greater than
// `value` or equal.
func GtOrEq(col SchemaField, value interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.GtOrEq{col.QualifiedName(schema): value}
	}
}

// Neq returns a condition that will be true when `col` is not `value`.
func Neq(col SchemaField, value interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.NotEq{col.QualifiedName(schema): value}
	}
}

// Or returns the given conditions joined by logical ors.
func Or(conds ...Condition) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.Or(condsToSqlizers(conds, schema))
	}
}

// And returns the given conditions joined by logical ands.
func And(conds ...Condition) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.And(condsToSqlizers(conds, schema))
	}
}

// Not returns the given condition negated.
func Not(cond Condition) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return not{cond(schema)}
	}
}

// In returns a condition that will be true when `col` is equal to any of the
// passed `values`.
func In(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.Eq{col.QualifiedName(schema): values}
	}
}

// NotIn returns a condition that will be true when `col` is distinct to all of the
// passed `values`.
func NotIn(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.NotEq{col.QualifiedName(schema): values}
	}
}

// ArrayEq returns a condition that will be true when `col` is equal to an
// array with the given elements.
func ArrayEq(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "=", types.Slice(values)}
	}
}

// ArrayNotEq returns a condition that will be true when `col` is not equal to
// an array with the given elements.
func ArrayNotEq(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "<>", types.Slice(values)}
	}
}

// ArrayLt returns a condition that will be true when all elements in `col`
// are lower or equal than their counterparts in the given values, and one of
// the elements at least is lower than its counterpart in the given values.
// For example: for a col with values [1,2,2] and values [1,2,3], it will be
// true.
func ArrayLt(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "<", types.Slice(values)}
	}
}

// ArrayGt returns a condition that will be true when all elements in `col`
// are greater or equal than their counterparts in the given values, and one of
// the elements at least is greater than its counterpart in the given values.
// For example: for a col with values [1,2,3] and values [1,2,2], it will be
// true.
func ArrayGt(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), ">", types.Slice(values)}
	}
}

// ArrayLtOrEq returns a condition that will be true when all elements in `col`
// are lower or equal than their counterparts in the given values.
// For example: for a col with values [1,2,2] and values [1,2,2], it will be
// true.
func ArrayLtOrEq(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "<=", types.Slice(values)}
	}
}

// ArrayGtOrEq returns a condition that will be true when all elements in `col`
// are greater or equal than their counterparts in the given values.
// For example: for a col with values [1,2,2] and values [1,2,2], it will be
// true.
func ArrayGtOrEq(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), ">=", types.Slice(values)}
	}
}

// ArrayContains returns a condition that will be true when `col` contains all the
// given values.
func ArrayContains(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "@>", types.Slice(values)}
	}
}

// ArrayContainedBy returns a condition that will be true when `col` has all
// its elements present in the given values.
func ArrayContainedBy(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "<@", types.Slice(values)}
	}
}

// ArrayOverlap returns a condition that will be true when `col` has elements
// in common with an array formed by the given values.
func ArrayOverlap(col SchemaField, values ...interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "&&", types.Slice(values)}
	}
}

// JSONIsObject returns a condition that will be true when `col` is a JSON
// object.
func JSONIsObject(col SchemaField) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colUnaryOp{col.QualifiedName(schema), " @> '{}'"}
	}
}

// JSONIsArray returns a condition that will be true when `col` is a JSON
// array.
func JSONIsArray(col SchemaField) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colUnaryOp{col.QualifiedName(schema), " @> '[]'"}
	}
}

// JSONContains returns a condition that will be true when `col` contains
// the given element converted to JSON.
func JSONContains(col SchemaField, elem interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "@>", types.JSON(elem)}
	}
}

// JSONContainsAny returns a condition that will be true when `col` contains
// any of the given elements converted to json.
// Giving no elements will cause an error to be returned when the condition is
// evaluated.
func JSONContainsAny(col SchemaField, elems ...interface{}) Condition {
	if len(elems) == 1 {
		return JSONContains(col, elems[0])
	}
	return func(schema Schema) squirrel.Sqlizer {
		if len(elems) == 0 {
			return &errOp{"can't check if json contains 0 elements"}
		}
		return &containsAny{col.QualifiedName(schema), elems}
	}
}

// JSONContainedBy returns a condition that will be true when `col` is
// contained by the given element converted to JSON.
func JSONContainedBy(col SchemaField, elem interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "<@", types.JSON(elem)}
	}
}

// JSONContainsAnyKey returns a condition that will be true when `col` contains
// any of the given keys. Will also match elements if the column is an array.
func JSONContainsAnyKey(col SchemaField, keys ...string) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "??|", types.Slice(keys)}
	}
}

// JSONContainsAllKeys returns a condition that will be true when `col`
// contains all the given keys. Will also match elements if the column is an
// array.
func JSONContainsAllKeys(col SchemaField, keys ...string) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return &colOp{col.QualifiedName(schema), "??&", types.Slice(keys)}
	}
}

type (
	not struct {
		cond squirrel.Sqlizer
	}

	colOp struct {
		col    string
		op     string
		valuer driver.Valuer
	}

	colUnaryOp struct {
		col string
		op  string
	}

	errOp struct {
		msg string
	}

	containsAny struct {
		col    string
		values []interface{}
	}
)

func (n not) ToSql() (string, []interface{}, error) {
	sql, args, err := n.cond.ToSql()
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("NOT (%s)", sql), args, err
}

func (o colOp) ToSql() (string, []interface{}, error) {
	return fmt.Sprintf("%s %s ?", o.col, o.op), []interface{}{o.valuer}, nil
}

func (o colUnaryOp) ToSql() (string, []interface{}, error) {
	return fmt.Sprintf("%s %s", o.col, o.op), nil, nil
}

func (o errOp) ToSql() (string, []interface{}, error) {
	return "", nil, errors.New(o.msg)
}

func (o containsAny) ToSql() (string, []interface{}, error) {
	var placeholders = make([]string, len(o.values))
	var args = make([]interface{}, len(o.values))
	for i, el := range o.values {
		args[i] = types.JSON(el)
		placeholders[i] = "?"
	}
	return fmt.Sprintf(
		"%s @> ANY (ARRAY [%s]::jsonb[])",
		o.col,
		strings.Join(placeholders, ", "),
	), args, nil
}

func condsToSqlizers(conds []Condition, schema Schema) []squirrel.Sqlizer {
	var result = make([]squirrel.Sqlizer, len(conds))
	for i, v := range conds {
		result[i] = v(schema)
	}
	return result
}
