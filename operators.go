package kallax

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

// Condition represents a condition of filtering in a query.
type Condition func(Schema) squirrel.Sqlizer

type not struct {
	cond squirrel.Sqlizer
}

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
func Gt(col SchemaField, n interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.Gt{col.QualifiedName(schema): n}
	}
}

// LtOrEq returns a condition that will be true when `col` is lower than
// `value` or equal.
func LtOrEq(col SchemaField, n interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.LtOrEq{col.QualifiedName(schema): n}
	}
}

// GtOrEq returns a condition that will be true when `col` is greater than
// `value` or equal.
func GtOrEq(col SchemaField, n interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.GtOrEq{col.QualifiedName(schema): n}
	}
}

// Neq returns a condition that will be true when `col` is not `value`.
func Neq(col SchemaField, n interface{}) Condition {
	return func(schema Schema) squirrel.Sqlizer {
		return squirrel.NotEq{col.QualifiedName(schema): n}
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

func condsToSqlizers(conds []Condition, schema Schema) []squirrel.Sqlizer {
	var result = make([]squirrel.Sqlizer, len(conds))
	for i, v := range conds {
		result[i] = v(schema)
	}
	return result
}

func (n not) ToSql() (string, []interface{}, error) {
	sql, args, err := n.cond.ToSql()
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("NOT (%s)", sql), args, err
}
