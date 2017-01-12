package operators

import (
	"gopkg.in/mgo.v2/bson"
)

// Or Joins query clauses with a logical OR returns all documents that match
// the conditions of either clause.
func Or(clauses ...bson.M) bson.M {
	return bson.M{"$or": clauses}
}

// And Joins query clauses with a logical AND returns all documents that match
// the conditions of both clauses.
func And(clauses ...bson.M) bson.M {
	return bson.M{"$and": clauses}
}

// Not Inverts the effect of a query expression and returns documents that do
// not match the query expression.
func Not(expr bson.M) bson.M {
	result := make(bson.M, 0)
	for key, e := range expr {
		result[key] = bson.M{"$not": e}
	}

	return result
}

// Nor Joins query clauses with a logical NOR returns all documents that fail
// to match both clauses.
func Nor(clauses ...bson.M) bson.M {
	return bson.M{"$nor": clauses}
}
