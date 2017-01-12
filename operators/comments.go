package operators

import (
	"gopkg.in/mgo.v2/bson"
)

// Comment Adds a comment to a query predicate.
func Comment(comment string) bson.M {
	return bson.M{"$comment": comment}
}
